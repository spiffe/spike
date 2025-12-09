![SPIKE](../assets/spike-banner-lg.png)

## SPIKE Bootstrapping Flow

**SPIKE Bootstrap** is responsible for initial system setup. It generates the 
root key, creates Shamir shards, and distributes them to **SPIKE Keeper**s.

**Kubernetes vs. Bare-Metal Differences:**

| Aspect             | Kubernetes             | Bare-Metal              |
|--------------------|------------------------|-------------------------|
| Bootstrap check    | ConfigMap flag         | Always proceeds         |
| Marker persistence | ConfigMap in namespace | None (runs each time)   |
| Keeper discovery   | K8s Service DNS        | Static URLs from config |
| Deployment         | DaemonSet/Deployment   | Systemd service         |

**Configuration:**
* `SPIKE_SHAMIR_THRESHOLD`: Number of shards needed (default: 2)
* `SPIKE_SHAMIR_SHARES`: Total shards to generate (default: 4)
* `SPIKE_KEEPERS`: Comma-separated Keeper URLs
* `SPIKE_NEXUS_URL`: SPIKE Nexus endpoint
* `SPIKE_BOOTSTRAP_FORCE`: Force bootstrap even if already completed (K8s only)

**Security Notes:**
* Root key NEVER persisted to disk
* Shards distributed immediately after generation
* Verification ensures correct key distribution
* Fail-fast on any security violation
* All sensitive data zeroed after use

**Root Key Generation Details:**

**1. Generate a 32-byte cryptographically secure random seed**
* Uses crypto/rand.Read()
* Entropy from OS

**2. Create P256 elliptic curve scalar**
* Deterministic operations from seed
* Scalar is the root key

**3. Create Shamir shares**
* Deterministic reader seeded with the root key
* Ensures consistent shards after restart
* threshold t = ShamirThreshold - 1
* total shares n = ShamirShares

**4. Verify shares**
* Take any t+1 shares
* Reconstruct secret
* Compare with original
* Fail-fast if mismatch

```mermaid
sequenceDiagram
    participant Bootstrap as SPIKE Bootstrap
    participant Lifecycle as Lifecycle<br/>Manager
    participant ConfigMap as ConfigMap<br/>(K8s only)
    participant State as State<br/>Manager
    participant Crypto as Crypto<br/>Primitives
    participant Shamir as Shamir Secret<br/>Sharing
    participant Keepers as SPIKE Keepers<br/>(all instances)
    participant Nexus as SPIKE Nexus
    participant Verify as Verification<br/>Logic

    Note over Bootstrap: SPIKE Bootstrap starts

    Bootstrap->>Lifecycle: ShouldBootstrap()

    alt Kubernetes
        Lifecycle->>ConfigMap: Check if bootstrapped
        ConfigMap-->>Lifecycle: bootstrapped flag
        alt Already bootstrapped
            Lifecycle-->>Bootstrap: false (skip bootstrap)
            Note over Bootstrap: Exit gracefully
        else Not bootstrapped
            Lifecycle-->>Bootstrap: true (proceed)
        end
    else Bare-Metal
        Lifecycle->>Lifecycle: Not in Kubernetes cluster
        Lifecycle-->>Bootstrap: true (always proceed)
    end

    Note over Bootstrap: Generate root key and shards

    Bootstrap->>State: RootShares()

    State->>Crypto: Generate 32-byte random seed
    Note right of Crypto: crypto/rand.Read()
    Crypto-->>State: []byte seed

    State->>State: Create P256 scalar from seed
    Note right of State: Elliptic curve scalar<br/>deterministic operations

    State->>Shamir: shamir.New(reader, t, secret)
    Note right of Shamir: t = ShamirThreshold - 1<br/>Example: t=2 (need 3 shards)

    Shamir->>Shamir: Split secret into n shares
    Note right of Shamir: n = ShamirShares<br/>Example: n=5 (5 total shards)

    Shamir-->>State: []ShamirShard

    State->>State: Verify reconstruction
    Note right of State: Take t+1 shards<br/>Reconstruct and compare<br/>Fail-fast if mismatch

    State-->>Bootstrap: rootKey, shards

    Note over Bootstrap,Keepers: Broadcast shards to all Keepers

    par Broadcast to each Keeper
        Bootstrap->>Keepers: POST /v1/store/contribute<br/>{shard: ShamirShard}
        Note right of Bootstrap: mTLS with SVID<br/>SPIFFE ID: spike/bootstrap

        Keepers->>Keepers: Validate peer is Bootstrap
        Keepers->>Keepers: state.SetShard(shard)
        Note right of Keepers: Store in memory<br/>Protected by mutex

        Keepers-->>Bootstrap: 200 OK
    end

    alt All Keepers acknowledged
        Note over Bootstrap: Verify initialization

        Bootstrap->>Verify: VerifyInitialization()

        Verify->>Crypto: Generate random 32-byte plaintext
        Crypto-->>Verify: plaintext

        Verify->>Crypto: Encrypt with root key
        Note right of Crypto: AES-256-GCM<br/>Generate 12-byte nonce
        Crypto-->>Verify: nonce, ciphertext

        Verify->>Nexus: POST /v1/bootstrap/verify<br/>{plaintext_hash, nonce, ciphertext}
        Note right of Verify: mTLS with SVID

        Nexus->>Nexus: Wait for Keepers to send shards
        Note right of Nexus: Collect threshold shards<br/>Reconstruct root key

        Nexus->>Nexus: Decrypt ciphertext with root key
        Nexus->>Nexus: Hash plaintext and compare

        alt Hash matches
            Nexus-->>Verify: 200 OK {success: true}

            Verify-->>Bootstrap: Verification successful

            opt Kubernetes only
                Bootstrap->>ConfigMap: Mark bootstrapped
                ConfigMap-->>Bootstrap: Updated
            end

            Note over Bootstrap: Bootstrap complete.<br/>System operational.
        else Hash mismatch
            Nexus-->>Verify: 500 Error
            Verify-->>Bootstrap: Verification failed
            Bootstrap->>Bootstrap: log.FatalErr()
            Note over Bootstrap: Abort bootstrap<br/>Security violation
        end
    else Some Keepers failed
        Bootstrap->>Bootstrap: log.FatalErr()
        Note over Bootstrap: Cannot proceed<br/>Need all Keepers online
    end
```