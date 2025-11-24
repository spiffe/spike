+++
#    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
#  \\\\\ Copyright 2024-present SPIKE contributors.
# \\\\\\\ SPDX-License-Identifier: Apache-2.0

title = "Nil Source Handling and Go Concurrency Guarantees"
weight = 100
+++

# Nil Source Handling and Go Concurrency Guarantees

This document explains the defensive nil checks for `*workloadapi.X509Source`
throughout the SPIKE codebase and the concurrency guarantees that make them
safe.

## Table of Contents

1. [Why Check for Nil Source?](#why-check-for-nil-source)
2. [When to Crash vs. Retry](#when-to-crash-vs-retry)
3. [Call Chain Analysis](#call-chain-analysis)
4. [Pointer Stability in Closures](#pointer-stability-in-closures)
5. [Go Error Handling Contract](#go-error-handling-contract)
6. [Concurrency and Memory Model](#concurrency-and-memory-model)

## Why Check for Nil Source?

In concurrent/distributed systems, the SPIFFE Workload API can asynchronously
invalidate the X509Source. While rare, this can happen due to:

- Workload API connection loss
- UDS socket issues
- SPIRE Agent problems
- Transient network failures

### Pattern Throughout Codebase

We implement defensive nil checks at strategic points to handle this gracefully.

## When to Crash vs. Retry

### Crash Immediately (Server Startup)

**Files**: `app/nexus/internal/net/serve.go`, `app/keeper/internal/net/serve.go`

```go
func Serve(appName string, source *workloadapi.X509Source) {
    // Fail-fast if source is nil: server cannot operate without mTLS
    if source == nil {
        log.FatalLn(
            appName,
            "message", "X509 source is nil, cannot start TLS server",
        )
    }
    // ... start server
}
```

**Rationale**:
- **No retry mechanism**: This is one-time server startup
- **Cannot function**: Server requires mTLS to operate
- **Fail-fast is better**: Makes initialization problems immediately obvious
- **Configuration error**: If nil at startup, it's a bug that needs immediate
  attention

### Retry Gracefully (Periodic/Recovery Functions)

**Files**:
- `app/nexus/internal/initialization/recovery/recovery.go`
- `app/spike/internal/cmd/secret/put.go`

```go
func InitializeBackingStoreFromKeepers(source *workloadapi.X509Source) {
    // ...
    _, err := retry.Forever(ctx, func() (bool, *sdkErrors.SDKError) {
        // Early check: avoid unnecessary function call if source is nil
        if source == nil {
            log.Warn(fName, "message", "X509 source is nil, will retry")
            return false, sdkErrors.ErrRecoveryRetryFailed
        }
        // ... attempt recovery
    })
}
```

```go
func SendShardsPeriodically(source *workloadapi.X509Source) {
    for range ticker.C {
        // Early check: skip if source is nil
        if source == nil {
            log.Warn(fName, "message", "X509 source is nil: skipping shard send")
            continue
        }
        // ... send shards
    }
}
```

**Rationale**:
- **Retry mechanism exists**: Function runs periodically or in retry loop
- **Transient failures**: Source might become available later
- **Graceful degradation**: Log warning and try again
- **System resilience**: Allows recovery from temporary Workload API issues

## Call Chain Analysis

When source is used for server creation, it goes through several layers:

```
Serve(source)
  ↓
net.Serve(source, handler, port)
  ↓
CreateMTLSServerWithPredicate(source, port, predicate)
  ↓ (also checks nil and crashes)
MTLSServerConfig(source, source, authorizer)
  ↓
HookMTLSServerConfig(config, svid, bundle, authorizer)
  ↓
config.GetCertificate = GetCertificate(svid, opts...)      ← closure captures source
config.VerifyPeerCertificate = WrapVerifyPeerCertificate(..., bundle, ...) ← closure captures source
```

### Defense in Depth

1. **First check**: `Serve()` crashes with clear context-specific error
2. **Second check**: `CreateMTLSServerWithPredicate()` has SDK-level validation
3. **If both missed**: TLS callbacks would crash during handshake (unclear error)

**Why fail-fast is better**:
- Clear error at startup: "X509 source is nil, cannot start TLS server"
- vs. cryptic panic during first TLS handshake: stack trace with unclear context

## Pointer Stability in Closures

### How Closures Capture Source

When we create the TLS configuration, closures capture the source pointer:

```go
func GetCertificate(source *workloadapi.X509Source, opts ...Option) func(*tls.ClientHelloInfo) (*tls.Certificate, error) {
    // source is captured by the closure below
    // The closure captures the pointer VALUE (memory address)

    return func(hello *tls.ClientHelloInfo) (*tls.Certificate, error) {
        // source can never become nil here because:
        // 1. The closure captured the pointer value when GetCertificate was called
        // 2. That captured pointer value is immutable within this closure
        // 3. Even if someone did `source = nil` elsewhere, this closure's
        //    copy of the pointer is unaffected

        svid, err := source.GetX509SVID()  // Uses the captured pointer
        if err != nil {
            return nil, err
        }
        return &svid.Certificates[0], nil
    }
}
```

### Key Insights

1. ✅ `source` points to a concrete initialized struct
2. ✅ The closure captures that pointer value
3. ✅ Inside the closure, `source` can never be nil (unless it was nil when
   captured)
4. ✅ The pointer value is "frozen" in the closure at capture time

**This is why our nil check at server startup is critical:**
- If `source` is nil when we create server → closures capture nil → TLS
  handshakes crash
- If `source` is valid when we create server → closures capture valid pointer
  → TLS handshakes work
- Once captured, the pointer in the closure won't change to nil later

## Go Error Handling Contract

### The Guarantee Pattern

When calling `GetX509SVID()`, the function contract guarantees safety:

```go
func (s *X509Source) GetX509SVID() (*x509svid.SVID, error) {
    if err := s.checkClosed(); err != nil {
        return nil, err
    }

    s.mtx.RLock()
    svid := s.svid
    s.mtx.RUnlock()

    if svid == nil {
        // Defensive check - should be unreachable
        return nil, wrapX509sourceErr(errors.New("missing X509-SVID"))
    }
    return svid, nil
}
```

**Three possible return paths:**
1. **Closed source**: `return nil, err`
2. **Missing SVID**: `return nil, error`
3. **Success**: `return svid, nil` (guaranteed non-nil)

**Usage pattern:**
```go
svid, err := source.GetX509SVID()
if err != nil {
    return nil, err  // Don't use svid!
}
// If we're here, err == nil, which means svid is GUARANTEED to be non-nil
use(svid)  // Safe!
```

**The guarantee**: You can never get `(nil, nil)` from this function. You
either get:
- `(nil, error)` - don't use the SVID
- `(*x509svid.SVID, nil)` - SVID is guaranteed non-nil

## Concurrency and Memory Model

### The Question

Can `svid` become nil between the check and the return?

```go
if svid == nil {
    return nil, wrapX509sourceErr(errors.New("missing X509-SVID"))
}
// ← Could another goroutine set svid = nil here?
return svid, nil
```

### The Answer: No

**Because of how the code is structured:**

```go
s.mtx.RLock()
svid := s.svid  // ← Creates a LOCAL COPY of the pointer
s.mtx.RUnlock()

// From here on, `svid` is a local variable on this goroutine's stack
if svid == nil {
    return nil, wrapX509sourceErr(errors.New("missing X509-SVID"))
}
return svid, nil  // Returning the local copy
```

### Why `svid` Cannot Become Nil

1. **`svid` is a local variable** (stack-allocated, goroutine-private)
2. **It's a copy of the pointer value** made while holding the read lock
3. **Other goroutines can modify `s.svid`** (the struct field), but **NOT this
   local `svid` variable**
4. **Go's memory model guarantees** that local variables are private to the
   goroutine's stack frame

### Memory Model Diagram

```
Heap:                          Stack (your goroutine):
┌─────────────────┐           ┌──────────────────┐
│  SVID struct    │ ←─────────│ svid (pointer)   │ (local copy)
│  {data...}      │ ↖         └──────────────────┘
└─────────────────┘  │
                     │
X509Source:          │
┌─────────────────┐  │
│ svid (pointer)  │──┘ (can be modified by other goroutines)
└─────────────────┘
```

**What's happening:**

1. ✅ There's a concrete `SVID` struct somewhere in memory
2. ✅ `s.svid` is already a pointer to that struct (type `*x509svid.SVID`)
3. ✅ `svid := s.svid` **copies the pointer value** (the address)
4. ✅ Now you have a **local copy of the address** (the pointer value)
5. ✅ Both `s.svid` and your local `svid` point to the **same concrete struct**
6. ✅ GC won't deallocate the struct because there's at least one reference to
   it
7. ✅ Your local `svid` is stack-local - other goroutines can't modify it
8. ✅ You return that pointer, guaranteed to be non-nil (if it passed the check)

### Go's Guarantees

- **Local variables are goroutine-private**: Each goroutine has its own stack
- **Copying a pointer copies the address value**: Atomic on all Go architectures
- **Other goroutines can't modify your stack frame**: Memory isolation
- **The pointer value is immutable after the copy**: Unless you reassign it in
  the same goroutine

### The Incorrect Pattern (Don't Do This)

```go
// UNSAFE - DON'T DO THIS
s.mtx.RLock()
if s.svid == nil {  // Checking the shared field
    s.mtx.RUnlock()
    return nil, error
}
// ← Another goroutine could set s.svid = nil here!
result := s.svid
s.mtx.RUnlock()
return result, nil  // Could return nil!
```

### The Correct Pattern (What It Actually Does)

```go
// SAFE - this is what it actually does
s.mtx.RLock()
svid := s.svid  // Atomic copy under lock
s.mtx.RUnlock()
// Now working with local copy - other goroutines can't touch it
if svid == nil {
    return nil, error
}
return svid, nil  // Guaranteed non-nil
```

## Summary

### Nil Source Handling Strategy

| Context | Strategy | Rationale |
|---------|----------|-----------|
| **Server Startup** | Crash with `log.FatalLn` | No retry mechanism; fail-fast makes problems obvious |
| **Retry Loops** | Warn and retry | Transient failures; source might recover |
| **Periodic Functions** | Warn and skip iteration | Source might become available in next cycle |
| **CLI Commands** | Print error and exit cleanly | User-facing; clear error message |

### Key Takeaways

1. **Defense in depth**: Multiple layers of nil checks throughout the call chain
2. **Fail-fast at startup**: Crash immediately if server can't function
3. **Graceful degradation in loops**: Warn and retry for transient failures
4. **Pointer stability**: Closures capture pointer values safely
5. **Go's error contract**: If `err == nil`, return value is guaranteed valid
6. **Memory model**: Local variables are goroutine-private
7. **Atomic pointer copies**: Safe under mutex protection

### Final Wisdom

**Check nil at startup, crash early, sleep well at night!** ✓

The combination of defensive nil checks, proper error handling, and Go's memory
model guarantees ensures that SPIKE handles source invalidation safely across
all operational contexts.