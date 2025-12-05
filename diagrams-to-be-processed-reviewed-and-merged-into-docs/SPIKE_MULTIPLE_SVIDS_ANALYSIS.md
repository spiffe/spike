# SPIKE Multiple SVIDs Analysis and Trust Model

**Date:** 2025-12-05
**Discussion Summary:** Comprehensive analysis of how SPIKE handles workloads
with multiple SPIFFE IDs (SVIDs), including security implications, trust
boundaries, and SPIFFE standards compliance.

---

## Table of Contents

1. [Executive Summary](#executive-summary)
2. [Initial Question and Context](#initial-question-and-context)
3. [SPIFFE X.509-SVID Standard](#spiffe-x509-svid-standard)
4. [Multiple SVIDs Per Workload](#multiple-svids-per-workload)
5. [How SPIKE Handles Authentication](#how-spike-handles-authentication)
6. [The PeerCertificates Array](#the-peercertificates-array)
7. [Security Analysis](#security-analysis)
8. [Trust Model](#trust-model)
9. [Operational Guidance](#operational-guidance)
10. [Recommendations](#recommendations)
11. [Jira Entries Created](#jira-entries-created)

---

## Executive Summary

**Question:** Can a workload have more than one SPIFFE ID/SVID? If so, how
does SPIKE behave? Is there a security gap?

**Answer:**
- ✅ **YES**, workloads can have multiple SVIDs (multiple separate certificates)
- ✅ SPIKE correctly validates and trusts presented SVIDs per SPIFFE standard
- ✅ **NO security gap** - This is working as designed
- ⚠️ Requires proper SPIRE configuration and operator understanding
- 📝 **Recommendation:** Add documentation and demo workload

---

## Initial Question and Context

### The Question

> "Related: can a workload have more than one SPIFFEID / SVID? If so how would
> SPIKE behave? pick up and verify the first one it finds? if so, that might
> be a gap too."

### Important Clarification

During the discussion, we clarified two related but distinct concepts:

1. **Multiple URI SANs in ONE certificate** - NOT allowed by SPIFFE
2. **Multiple separate SVID certificates for ONE workload** - Allowed and
   supported

This document addresses #2, which is the valid and interesting case.

---

## SPIFFE X.509-SVID Standard

### Single SPIFFE ID Per Certificate

**From the SPIFFE X.509-SVID specification:**

> "An X.509 SVID MUST contain exactly one URI SAN, and by extension, exactly
> one SPIFFE ID."

> "SVIDs containing more than one URI SAN MUST be rejected."

**Rationale from the spec:**
- Multiple SPIFFE IDs create auditing challenges
- Multiple SPIFFE IDs complicate authorization logic
- Multiple URI SANs create validation challenges

### What IS Allowed

```
✅ Exactly ONE URI SAN (the SPIFFE ID)
✅ Multiple other SAN types (DNS SANs, email SANs, etc.)
✅ Multiple SEPARATE X.509-SVID certificates for same workload
```

### go-spiffe Library Enforcement

The `IDFromCert()` function enforces this:

```go
func IDFromCert(cert *x509.Certificate) (spiffeid.ID, error)
```

**Documentation:**
> "IDFromCert extracts the SPIFFE ID from the URI SAN of the provided
> certificate. It will return an error if the certificate does not have
> exactly one URI SAN with a well-formed SPIFFE ID."

**Error cases:**
1. Certificate has **zero** URI SANs → Error
2. Certificate has **more than one** URI SAN → **Error** ❌
3. URI SAN is not a valid SPIFFE ID → Error

**Key point:** The library automatically rejects certificates with multiple
URI SANs, so SPIKE inherits this protection.

---

## Multiple SVIDs Per Workload

### How It Works

A workload CAN receive multiple SEPARATE X.509-SVID certificates from SPIRE.

**Example SPIRE Registration:**

```bash
# Registration Entry 1 - Internal identity
spire-server entry create \
  -parentID spiffe://example.org/spire/agent/node1 \
  -spiffeID spiffe://example.org/app/internal \
  -selector unix:uid:1000 \
  -hint internal

# Registration Entry 2 - External identity
spire-server entry create \
  -parentID spiffe://example.org/spire/agent/node1 \
  -spiffeID spiffe://example.org/app/external \
  -selector unix:uid:1000 \
  -hint external

# Registration Entry 3 - Admin identity (optional)
spire-server entry create \
  -parentID spiffe://example.org/spire/agent/node1 \
  -spiffeID spiffe://example.org/app/admin \
  -selector unix:uid:1000 \
  -hint admin
```

**What happens:**
1. Workload with UID 1000 matches **all three** registration entries
2. SPIRE Workload API returns **three X.509-SVID certificates**
3. Each certificate contains **one SPIFFE ID** (in its URI SAN)
4. Workload receives all three but uses them separately

### The Hint Field

**From SPIFFE Workload API specification:**

The `hint` field is an optional string in the X509SVID protobuf message:

> "An operator-specified string used to provide guidance on how this identity
> should be used by a workload when more than one SVID is returned. For
> example, `internal` and `external` to indicate an SVID for internal or
> external use, respectively."

**Constraints:**
- When set (non-empty), SPIFFE Workload API servers MUST ensure hint values
  are unique within each X509SVIDResponse
- If duplicates occur, "the first message in the list SHOULD be selected"
- Workloads are responsible for handling missing or unexpected hint values

**Purpose:**
- Guide workload behavior
- Document identity purpose
- Help workload choose which SVID to use for which connection
- **NOT for access control** (that's the policy engine's job)

### SVID Selection by Workload

**Default behavior (go-spiffe):**

```go
source, _ := workloadapi.NewX509Source(ctx)
svid := source.GetX509SVID()  // Returns FIRST SVID in list
```

The SPIFFE specification states: "The default identity is the first in the
`svids` list returned in the `X509SVIDResponse` message."

**Custom selection using hint:**

```go
source, _ := workloadapi.NewX509Source(ctx,
    workloadapi.WithDefaultX509SVIDPicker(func(svids []*x509svid.SVID) *x509svid.SVID {
        // Custom logic to select SVID
        for _, s := range svids {
            if s.Hint == "external" {
                log.Printf("Selected SVID %s for external use", s.ID)
                return s  // Use "external" SVID for this connection
            }
        }
        // Fallback to first SVID
        return svids[0]
    }),
)
```

**Key insight:** The **workload** chooses which SVID to present. SPIKE (or any
server) has no control over this choice.

---

## How SPIKE Handles Authentication

### Authentication Flow

**Code path:** `internal/auth/spiffe.go`

```go
func ExtractPeerSPIFFEID[T any](
    r *http.Request,
    w http.ResponseWriter,
    errorResponse T,
) (*spiffeid.ID, *sdkErrors.SDKError) {
    // Step 1: Extract SPIFFE ID from TLS certificate
    peerSPIFFEID, err := spiffe.IDFromRequest(r)  // Uses SDK
    if err != nil {
        // Certificate invalid or doesn't have exactly one URI SAN
        return nil, sdkErrors.ErrAccessUnauthorized.Wrap(...)
    }

    // Step 2: Additional validation
    err = validation.ValidateSPIFFEID(peerSPIFFEID.String())
    if err != nil {
        // SPIFFE ID format invalid
        return nil, sdkErrors.ErrAccessUnauthorized.Wrap(...)
    }

    return peerSPIFFEID, nil
}
```

**Then used in guards:**

```go
// app/nexus/internal/route/secret/guard.go
func guardSecretRequest(...) {
    // Extract SPIFFE ID
    peerSPIFFEID, err := auth.ExtractPeerSPIFFEID(r, w, ...)

    // Check access using extracted SPIFFE ID
    allowed := state.CheckAccess(peerSPIFFEID.String(), path, permissions)
    if !allowed {
        // 403 Forbidden
    }
}
```

### What SPIKE Sees

When a client connects:

```
1. Client chooses which SVID to present (before TLS handshake)
2. TLS handshake with chosen SVID
3. SPIKE receives mTLS connection
4. SPIKE extracts: peerSPIFFEID = spiffe.IDFromRequest(r)
5. SPIKE enforces policies based on that SPIFFE ID
```

**SPIKE has NO control over which SVID the client chose to present.**

### Multiple SVIDs Scenario

```
Workload has:
  - spiffe://example.org/app/lowprivilege (hint: "internal")
  - spiffe://example.org/app/admin (hint: "external")

Connection to SPIKE:
  1. Workload uses hint-based picker
  2. Selects "admin" SVID for this connection
  3. Presents admin certificate during TLS handshake
  4. SPIKE extracts: spiffe://example.org/app/admin
  5. SPIKE applies policies for "admin" identity
  6. If policies allow, grants access
```

---

## The PeerCertificates Array

### Important Clarification

During discussion, we clarified what `PeerCertificates` actually contains.

**Question:** "PeerCertificate is an array; does that imply the client can
send multiple certs, hence SVIDs?"

**Answer:** **NO** - PeerCertificates contains the certificate **chain**, not
multiple identities.

### Certificate Chain Structure

```
PeerCertificates[0] = Leaf certificate (the actual client SVID)
PeerCertificates[1] = Intermediate CA certificate
PeerCertificates[2] = Another intermediate CA (if any)
...
PeerCertificates[N] = Root CA (usually not sent, in trust store)
```

### Visual Example

```
┌─────────────────────────────────────┐
│ PeerCertificates[0]                 │
│ ─────────────────────────────────── │
│ Subject: CN=workload                │
│ URI SAN: spiffe://example.org/app   │ ← THE SVID (leaf cert)
│ Issued by: Intermediate CA          │
└─────────────────────────────────────┘
              ↓ signed by
┌─────────────────────────────────────┐
│ PeerCertificates[1]                 │
│ ─────────────────────────────────── │
│ Subject: CN=SPIRE Intermediate CA   │ ← Intermediate CA
│ Issued by: Root CA                  │
└─────────────────────────────────────┘
              ↓ signed by
┌─────────────────────────────────────┐
│ Root CA (in trust store)            │
│ ─────────────────────────────────── │
│ Subject: CN=SPIRE Root CA           │ ← Root CA
│ Self-signed                         │
└─────────────────────────────────────┘
```

### How go-spiffe Uses It

```go
func IDFromRequest(r *http.Request) (*spiffeid.ID, error) {
    cs := r.TLS  // TLS connection state

    // Extract from FIRST cert only (the leaf)
    return IDFromCert(cs.PeerCertificates[0])

    // PeerCertificates[1:] are CA certs, ignored for ID extraction
}
```

### TLS Protocol Limitation

**Key point:** In a single TLS handshake, a client can only present **one
certificate chain**.

This means:
- Client presents: **1 leaf cert** (1 SVID) + CA certificates
- If workload has multiple SVIDs, it **chooses one** before connecting
- Only **that one SVID** appears in `PeerCertificates[0]`
- Other SVIDs the workload might have are **not sent**

---

## Security Analysis

### Potential Attack Scenario

**Setup:**
```
SPIRE Registration:
  Entry 1: UID 1000 → spiffe://example.org/app/lowprivilege (hint: internal)
  Entry 2: UID 1000 → spiffe://example.org/app/admin (hint: external)

SPIKE Policies:
  Policy 1: lowprivilege → READ secrets/public/*
  Policy 2: admin → READ+WRITE secrets/sensitive/*
```

**Attack:**
```
1. Malicious workload (UID 1000) connects to SPIKE
2. Chooses to present the "admin" SVID
3. SPIKE sees spiffe://example.org/app/admin
4. Policy check passes (admin has access)
5. Workload accesses secrets/sensitive/*
```

**Question:** Is this a privilege escalation?

### Answer: NO - This is Working as Designed

**Why the attack scenario is NOT a security bug:**

1. **SPIRE authorized it:** SPIRE issued both SVIDs to this workload based on
   registration entries with matching selectors

2. **Operator configured it:** The operator explicitly created both
   registration entries with the same selectors (unix:uid:1000)

3. **SPIFFE trust model:** If SPIRE issued an SVID, the workload IS
   authorized for that identity

4. **Certificate is valid:** The presented certificate is cryptographically
   valid and chains to a trusted root

5. **SPIKE's role:** SPIKE enforces policies based on the presented identity,
   NOT on whether the workload "should" have that identity (that's SPIRE's
   job)

### Who Decides Which SPIFFE ID a Workload Can Use?

**Answer: SPIRE does** - through registration entries and selectors.

**SPIRE's responsibility:**
- Attest workload identity (via selectors: UID, K8s SA, etc.)
- Issue SVIDs only to workloads matching selectors
- Verify attestation before issuing certificates

**SPIKE's responsibility:**
- Validate presented certificate cryptographically
- Extract SPIFFE ID from certificate
- Enforce policies based on that SPIFFE ID

**Separation of concerns:**
- Identity issuance: SPIRE
- Authorization: SPIKE

### Is There a Gap?

**NO - But it depends on understanding the trust model.**

**SPIKE's trust model:**
```
IF:
  1. Certificate is cryptographically valid (signature, chain)
  2. Certificate chains to a trusted root (SPIRE)
  3. Certificate contains valid SPIFFE ID

THEN:
  Trust that SPIRE issued this SVID to an authorized workload
```

**SPIKE does NOT:**
- Query SPIRE to verify "should this workload have this SPIFFE ID?"
- Re-implement workload attestation
- Restrict which SVID a multi-SVID workload can use

**Why this is correct:**
- Standard SPIFFE/Zero Trust model
- Same approach as Vault, Istio, Envoy, etc.
- SPIRE is the trusted identity authority
- Separation of concerns

---

## Trust Model

### SPIKE's Security Relies On

1. **SPIRE correctly attesting workload identity** (via selectors)
2. **SPIRE only issuing SVIDs to authorized workloads**
3. **Cryptographic validation** of presented certificates
4. **Operators configuring SPIRE registration correctly**

### SPIKE Does NOT Validate

- Whether a workload "should" have a particular SPIFFE ID
- Whether SPIRE's selector configuration is correct
- Whether multiple SVIDs for same selectors is appropriate

**Analogy to other systems:**

| System | Trust Model |
|--------|-------------|
| Vault | Trusts Kubernetes service account tokens |
| Istio | Trusts SPIRE-issued SVIDs |
| AWS IAM | Trusts EC2 instance profiles |
| SPIKE | **Trusts SPIRE-issued SVIDs** |

All delegate identity verification to the authority.

### Defense-in-Depth Layers

Even with this trust model, there are multiple security layers:

1. ✅ **SPIRE attestation** - Workload must match selectors to get SVID
2. ✅ **TLS validation** - Certificate must be cryptographically valid
3. ✅ **Chain of trust** - Certificate must chain to trusted root
4. ✅ **go-spiffe enforcement** - Exactly one URI SAN required
5. ✅ **SPIKE validation** - Additional SPIFFE ID format checks
6. ✅ **Policy enforcement** - Access based on presented identity

---

## Operational Guidance

### The Real Risk: Misconfiguration

**The "risk" is operational, not technical:**

If an operator creates multiple registration entries with the same selectors
but different privilege levels, they've **intentionally** granted that workload
multiple identities.

**Example of misconfiguration:**

```bash
# WRONG: Same selectors for different privilege levels
spire-server entry create \
  -spiffeID spiffe://example.org/app/user \
  -selector unix:uid:1000

spire-server entry create \
  -spiffeID spiffe://example.org/app/admin \
  -selector unix:uid:1000  # ← Too broad! Same as above
```

Now **any** process with UID 1000 can use either identity.

### Best Practices

**✅ DO: Use different selectors for different privilege levels**

```bash
# Correct: Different selectors
spire-server entry create \
  -spiffeID spiffe://example.org/admin-service \
  -selector kubernetes:sa:admin-service

spire-server entry create \
  -spiffeID spiffe://example.org/user-service \
  -selector kubernetes:sa:user-service
```

**✅ DO: Use multiple SVIDs for legitimate use cases**

```bash
# Legitimate: Internal vs external routing
spire-server entry create \
  -spiffeID spiffe://example.org/gateway/internal \
  -selector kubernetes:sa:api-gateway \
  -hint internal

spire-server entry create \
  -spiffeID spiffe://example.org/gateway/external \
  -selector kubernetes:sa:api-gateway \
  -hint external
```

Use case: API gateway uses "internal" SVID for backend services, "external"
SVID for internet-facing APIs.

**❌ DON'T: Mix low/high privilege SVIDs with same selectors**

Unless there's a specific architectural reason (and it's well-documented).

**✅ DO: Use hints semantically**

```
Good hints: "internal", "external", "admin", "readonly"
Bad hints: "svid1", "cert-a", "production"
```

**✅ DO: Document why workload needs multiple identities**

In registration scripts, comments, or documentation.

**✅ DO: Audit SPIRE registration entries**

```bash
# List all entries
spire-server entry show

# Check for overlapping selectors
spire-server entry show | grep "Selector.*unix:uid:1000"
```

### When to Use Multiple SVIDs

**Legitimate use cases:**

1. **Internal/External Routing**
   - Service mesh: internal vs external traffic
   - API gateway: backend vs public APIs

2. **Multi-Tenant Applications**
   - Different identity per tenant
   - Isolation within single process

3. **Privilege Separation**
   - Separate identity for admin operations
   - But use different selectors or document extensively

4. **Gateway/Proxy Scenarios**
   - Different downstream services require different identities
   - Proxy presents appropriate identity per backend

**Not recommended:**

1. Workaround for bad architecture
2. Mixing fundamentally different privilege levels
3. When separate processes/containers would be clearer
4. Without clear documentation of why

---

## Recommendations

### 1. No Code Changes Needed ✅

**SPIKE is working correctly per SPIFFE standard.**

Current behavior:
- Validates certificates cryptographically
- Extracts SPIFFE ID from presented SVID
- Enforces policies based on that identity
- Trusts SPIRE for identity issuance

**This is the correct SPIFFE trust model.**

### 2. Add Documentation 📝

**Priority: LOW** (clarification, not a bug)

**Location:** `docs-src/content/operations/security.md` or similar

**Sections to add:**

1. **"SPIFFE Trust Model and Multiple SVIDs"**
   - Explain SPIKE trusts SPIRE for identity issuance
   - Describe how multiple SVIDs work
   - Clarify SPIKE's role vs SPIRE's role
   - Explain hint field purpose

2. **"Identity Verification and Trust Boundaries"**
   - Document that SPIKE extracts SPIFFE ID from client cert
   - Note that workload chooses which SVID to present
   - Explain trust boundary between SPIKE and SPIRE
   - Reference SPIFFE X.509-SVID standard

3. **"SPIRE Registration Best Practices"**
   - Use different selectors for different privilege levels
   - When to use multiple SVIDs
   - How to audit registration entries
   - Common pitfalls and misconfigurations

### 3. Create Advanced Demo Workload 💡

**Priority: MEDIUM** (educational value)

**Purpose:** Demonstrate multiple-SVID-per-workload feature

**Demo name:** "spike-demo-multi-identity"

**Components:**
- Go application using go-spiffe
- SPIRE registration scripts
- SPIKE policy configuration
- README with detailed explanation
- Docker/Kubernetes deployment manifests

**Demo scenario:** "Multi-Environment Service Router"

A workload that routes requests to internal/external services and needs
different identities:
- Internal requests → "internal" SVID
- External requests → "external" SVID
- Admin operations → "admin" SVID

**Educational value:**
- Shows hint field usage
- Demonstrates `WithDefaultX509SVIDPicker`
- Illustrates SPIKE/SPIRE trust boundaries
- Teaches proper SPIRE configuration

**See Jira entry for full implementation details.**

### 4. Optional: Add Debug Logging

Consider adding debug-level logging in SPIKE:

```go
// When extracting SPIFFE ID
log.Debug("Authenticated connection",
    "spiffe_id", peerSPIFFEID.String(),
    "remote_addr", r.RemoteAddr,
)
```

This helps operators understand which identity workloads are using.

### 5. Optional: SPIRE Integration (Future)

**Not recommended** unless specific compliance requirement.

Could add SPIRE Server API integration to:
- Query registration entries
- Cross-check if presented SVID matches expected selectors
- Add extra validation layer

**Pros:**
- Defense-in-depth
- Catches SPIRE misconfigurations

**Cons:**
- Couples SPIKE to SPIRE Server API
- Performance overhead (extra API call per request)
- Duplicates SPIRE's attestation logic
- Breaks if SPIRE registration changes
- Goes against SPIFFE separation of concerns
- Not standard practice

**Decision:** Document the trust model instead. Let operators understand that
SPIRE configuration is their responsibility.

---

## Jira Entries Created

### 1. Documentation Issue

**File:** `jira.xml`
**Section:** `<documentation>`
**Priority:** Low
**Category:** security-clarification

**Title:** "Document SPIFFE trust model: Workloads with multiple SVIDs"

**Summary:**
- Explains how multiple SVIDs work
- Clarifies SPIKE's trust model and boundaries
- Provides SPIRE registration best practices
- Notes this is NOT a security bug
- Includes action items for documentation

**Key points:**
- SPIKE trusts SPIRE for identity issuance
- Workload chooses which SVID to present
- SPIKE enforces policies based on presented identity
- Separation of concerns (SPIRE vs SPIKE)

### 2. Demo Application Issue

**File:** `jira.xml`
**Section:** `<examples>`
**Priority:** Medium
**Category:** demo-application

**Title:** "Create advanced demo workload showcasing multiple SVIDs with
hint-based selection"

**Summary:**
- Full implementation plan for demo workload
- "Multi-Environment Service Router" scenario
- Shows internal/external/admin identity usage
- Includes code structure, SPIRE setup, SPIKE policies
- README sections, testing plan, video walkthrough

**Components:**
1. Go application using go-spiffe
2. SPIRE registration scripts with hints
3. SPIKE policy configuration
4. Comprehensive README
5. Docker/Kubernetes manifests
6. Architecture documentation

**Educational value:**
- Multiple SVIDs are a feature, not a bug
- Hint field usage in practice
- go-spiffe `WithDefaultX509SVIDPicker`
- SPIKE/SPIRE trust boundaries
- Proper SPIRE selector configuration
- Policy design for multi-identity workloads

---

## Conclusion

### Summary

**Question:** Can workloads have multiple SVIDs? Is this a security gap in
SPIKE?

**Answer:**

1. ✅ **Yes**, workloads can have multiple SVIDs (multiple separate
   certificates)

2. ✅ **SPIKE handles this correctly** per SPIFFE standard
   - Validates presented certificate
   - Extracts SPIFFE ID
   - Enforces policies based on that identity

3. ❌ **No security gap exists**
   - Working as designed
   - Follows SPIFFE trust model
   - Industry standard practice

4. ⚠️ **Requires proper SPIRE configuration**
   - Use different selectors for different privilege levels
   - Don't mix low/high privilege with same selectors
   - Document why workloads need multiple identities

5. 📝 **Documentation needed**
   - Explain trust model
   - Show best practices
   - Prevent operator confusion

6. 💡 **Demo workload recommended**
   - Educational value
   - Shows real SPIFFE feature
   - Illustrates trust boundaries

### Key Takeaways

1. **Multiple SVIDs per workload is a SPIFFE feature**, not a bug or gap

2. **The hint field helps workloads choose** which SVID to use for which
   purpose

3. **SPIKE trusts whatever valid SVID is presented** - this is correct per
   SPIFFE trust model

4. **SPIRE is responsible for identity issuance**, SPIKE is responsible for
   authorization

5. **The "risk" is operational (misconfiguration)**, not technical (security
   bug)

6. **Operators must understand** the trust boundary between SPIRE and SPIKE

7. **Best practice:** Use different selectors for different privilege levels

8. **Documentation and education** will prevent confusion

### Final Assessment

**No code changes needed.** SPIKE is working correctly.

**Action items:**
1. Add documentation explaining trust model
2. Create demo workload (educational)
3. Consider debug logging (optional)

**This discussion should be reviewed** by security-conscious operators to
understand how SPIKE and SPIRE work together in a Zero Trust architecture.

---

## References

### SPIFFE Specifications

1. **X.509-SVID Standard**
   - https://github.com/spiffe/spiffe/blob/main/standards/X509-SVID.md
   - "An X.509 SVID MUST contain exactly one URI SAN"
   - "SVIDs containing more than one URI SAN MUST be rejected"

2. **SPIFFE Workload API**
   - https://github.com/spiffe/spiffe/blob/main/standards/SPIFFE_Workload_API.md
   - X509SVIDResponse can contain multiple X509SVID messages
   - Hint field: "operator-specified string to provide guidance"
   - "First SVID in list is default identity"

### go-spiffe Library

1. **IDFromCert function**
   - https://pkg.go.dev/github.com/spiffe/go-spiffe/v2/svid/x509svid
   - "Returns error if certificate does not have exactly one URI SAN"

2. **WithDefaultX509SVIDPicker**
   - https://pkg.go.dev/github.com/spiffe/go-spiffe/v2/workloadapi
   - Custom picker for selecting from multiple SVIDs

### SPIKE Codebase

1. **Authentication:** `internal/auth/spiffe.go:49-85`
   - `ExtractPeerSPIFFEID()` function
   - Uses `spiffe.IDFromRequest()`

2. **Authorization:** `app/nexus/internal/route/secret/guard.go:42-74`
   - `guardSecretRequest()` function
   - Calls `CheckAccess()` with extracted SPIFFE ID

3. **Policy Check:** `app/nexus/internal/state/base/policy.go:43-70`
   - `CheckAccess()` function
   - Enforces policies based on SPIFFE ID

---

**Document End**

---

**Reviewed By:** [To be filled after review]
**Date:** [To be filled]
**Approved:** [To be filled]
