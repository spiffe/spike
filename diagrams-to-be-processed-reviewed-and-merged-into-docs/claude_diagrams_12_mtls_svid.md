# mTLS Establishment Using SPIFFE SVID

## Overview

SPIKE uses SPIFFE (Secure Production Identity Framework For Everyone) for
workload identity. All communication between components uses mutual TLS (mTLS)
with X.509 SVIDs (SPIFFE Verifiable Identity Documents) as certificates.

---

## 1. SVID Acquisition from SPIRE

```mermaid
sequenceDiagram
    participant Workload as SPIKE Component<br/>(Nexus/Keeper/Pilot)
    participant Socket as Unix Domain Socket<br/>/run/spire/sockets/agent.sock
    participant Agent as SPIRE Agent
    participant Server as SPIRE Server
    participant CA as SPIRE CA

    Note over Workload: Component starts

    Workload->>Socket: Connect to Workload API
    Note right of Socket: SPIFFE_ENDPOINT_SOCKET<br/>unix:///run/spire/sockets/agent.sock

    Socket->>Agent: Connection established

    Workload->>Agent: FetchX509SVID()
    Note right of Workload: gRPC streaming call<br/>go-spiffe SDK

    Agent->>Agent: Attest workload
    Note right of Agent: PID, UID, Kubernetes selectors,<br/>etc.

    Agent->>Agent: Lookup registration entry
    Note right of Agent: Match selectors to SPIFFE ID

    alt Workload not registered
        Agent-->>Workload: Error: No registration entry
        Note over Workload: Cannot proceed without SVID
    else Workload registered
        Agent->>Agent: Check SVID cache

        alt SVID cached and valid
            Agent-->>Workload: Return cached SVID
        else SVID not cached or expired
            Agent->>Server: RequestX509SVID()
            Note right of Agent: Authenticate with node SVID

            Server->>Server: Validate node identity
            Server->>Server: Authorize SVID request

            Server->>CA: Sign CSR for workload
            Note right of CA: Trust domain CA<br/>Issues X.509 certificate

            CA-->>Server: Signed X.509 certificate

            Server-->>Agent: X509SVID + trust bundle

            Agent->>Agent: Cache SVID

            Agent-->>Workload: X509SVID + trust bundle
        end

        Note over Workload: SVID acquired.<br/>Contains:<br/>- X.509 certificate<br/>- Private key<br/>- SPIFFE ID in SAN<br/>- Trust bundle (CA certs)
    end

    Workload->>Workload: Create X509Source
    Note right of Workload: Manages SVID lifecycle<br/>Automatic rotation

    loop SVID Rotation (every ~1 hour)
        Agent->>Workload: Push updated SVID
        Note right of Agent: Before expiration

        Workload->>Workload: Update X509Source
        Note right of Workload: Seamless rotation<br/>No downtime
    end
```

**Key Files:**
- `internal/net/spiffe.go::Source()` - Acquire X509Source
- SPIRE Go SDK: `github.com/spiffe/go-spiffe/v2/workloadapi`

**SPIFFE ID Format:**
```
spiffe://<trust-domain>/<workload-path>

Examples:
spiffe://example.org/spike/nexus
spiffe://example.org/spike/keeper/0
spiffe://example.org/spike/pilot/role/superuser
spiffe://example.org/workload/api
```

---

## 2. mTLS Server Setup (SPIKE Nexus Example)

```mermaid
sequenceDiagram
    participant Main as Nexus Main
    participant SPIFFE as SPIFFE Module
    participant Source as X509Source
    participant TLS as TLS Listener
    participant HTTP as HTTP Server

    Note over Main: SPIKE Nexus starts

    Main->>SPIFFE: Acquire X509Source
    SPIFFE->>Source: workloadapi.NewX509Source(ctx)
    Note right of Source: Connects to SPIRE Agent<br/>Starts SVID rotation

    Source-->>SPIFFE: X509Source

    SPIFFE->>SPIFFE: Validate self SPIFFE ID
    Note right of SPIFFE: Must match expected pattern<br/>spiffeid.IsNexus(id)

    SPIFFE-->>Main: X509Source + selfSPIFFEID

    Main->>TLS: Create TLS listener
    Note right of TLS: net.Listen("tcp", ":8553")

    Main->>TLS: Configure TLS
    Note right of TLS: tlsConfig := &tls.Config{<br/>  GetCertificate: source.GetCertificate,<br/>  GetClientCAs: source.GetClientCAs,<br/>  ClientAuth: tls.RequireAndVerifyClientCert,<br/>  MinVersion: tls.VersionTLS13,<br/>}

    TLS->>TLS: Wrap listener with TLS
    Note right of TLS: tls.NewListener(listener, tlsConfig)

    Main->>HTTP: Start HTTP server
    Note right of HTTP: http.Serve(tlsListener, handler)

    Note over HTTP: Server ready.<br/>Accepts mTLS connections.<br/>Verifies client SVIDs.
```

**Key Configuration:**

```go
tlsConfig := &tls.Config{
    // Server certificate (Nexus SVID)
    GetCertificate: func(hello *tls.ClientHelloInfo) (*tls.Certificate, error) {
        return source.GetX509SVID()
    },

    // CA bundle for verifying client certificates
    GetClientCAs: func() (*x509.CertPool, error) {
        return source.GetX509BundleForTrustDomain(trustDomain)
    },

    // Require client certificate
    ClientAuth: tls.RequireAndVerifyClientCert,

    // TLS 1.3 only
    MinVersion: tls.VersionTLS13,
}
```

**Key Files:**
- `app/nexus/internal/net/serve.go::Serve()`
- `internal/net/tls.go` - TLS configuration

---

## 3. mTLS Client Setup (SPIKE Pilot Example)

```mermaid
sequenceDiagram
    participant Pilot as SPIKE Pilot
    participant SPIFFE as SPIFFE Module
    participant Source as X509Source
    participant Client as HTTP Client
    participant Nexus as SPIKE Nexus

    Note over Pilot: SPIKE Pilot starts

    Pilot->>SPIFFE: Acquire X509Source
    SPIFFE->>Source: workloadapi.NewX509Source(ctx)
    Source-->>SPIFFE: X509Source

    SPIFFE-->>Pilot: X509Source

    Pilot->>Client: Create mTLS HTTP client

    Client->>Client: Configure TLS
    Note right of Client: tlsConfig := &tls.Config{<br/>  GetClientCertificate: source.GetClientCert,<br/>  RootCAs: source.GetTrustBundle(),<br/>  MinVersion: tls.VersionTLS13,<br/>}

    Client->>Client: Create HTTP transport
    Note right of Client: transport := &http.Transport{<br/>  TLSClientConfig: tlsConfig,<br/>}

    Client->>Client: Create HTTP client
    Note right of Client: client := &http.Client{<br/>  Transport: transport,<br/>}

    Pilot->>Client: Make request to Nexus

    Client->>Nexus: HTTPS GET /v1/secret<br/>(TLS handshake)

    Note over Client,Nexus: mTLS Handshake

    Nexus->>Nexus: Verify client certificate
    Note right of Nexus: Check signature against CA<br/>Extract SPIFFE ID from SAN

    Nexus-->>Client: Certificate verified

    Client->>Client: Verify server certificate
    Note right of Client: Check signature against CA<br/>Verify SPIFFE ID matches expected

    Client-->>Pilot: Certificate verified

    Note over Client,Nexus: mTLS connection established

    Client->>Nexus: Send HTTP request

    Nexus->>Nexus: Extract SPIFFE ID from client cert
    Nexus->>Nexus: Authorize request based on SPIFFE ID

    Nexus-->>Client: HTTP response

    Client-->>Pilot: Response data
```

**Key Configuration:**

```go
tlsConfig := &tls.Config{
    // Client certificate (Pilot SVID)
    GetClientCertificate: func(info *tls.CertificateRequestInfo) (*tls.Certificate, error) {
        return source.GetX509SVID()
    },

    // CA bundle for verifying server certificate
    RootCAs: source.GetX509BundleForTrustDomain(trustDomain),

    // TLS 1.3 only
    MinVersion: tls.VersionTLS13,

    // Optionally verify server SPIFFE ID
    VerifyPeerCertificate: func(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error {
        // Custom verification logic
        return verifyServerSPIFFEID(verifiedChains)
    },
}
```

**Key Files:**
- `app/spike/internal/net/client.go` - mTLS client creation
- `internal/net/client.go::CreateMTLSClient()`

---

## 4. SPIFFE ID Extraction and Validation

```mermaid
sequenceDiagram
    participant Handler as HTTP Handler
    participant Request as HTTP Request
    participant TLS as TLS Connection
    participant Cert as Client Certificate
    participant SPIFFE as SPIFFE Module
    participant Validation as Validation Logic

    Note over Handler: Process incoming request

    Handler->>Request: Get TLS connection state
    Request->>TLS: r.TLS

    TLS-->>Handler: *tls.ConnectionState

    Handler->>TLS: Get peer certificates
    Note right of TLS: connState.PeerCertificates

    alt No client certificate
        TLS-->>Handler: Empty slice
        Handler->>Handler: Reject request (401)
    else Client certificate present
        TLS-->>Handler: []*x509.Certificate

        Handler->>Cert: Extract first certificate
        Note right of Cert: peerCerts[0]

        Handler->>SPIFFE: Extract SPIFFE ID
        SPIFFE->>Cert: Parse SAN (Subject Alternative Name)
        Note right of Cert: URI SAN:<br/>spiffe://example.org/spike/pilot/role/superuser

        alt No SPIFFE ID in SAN
            SPIFFE-->>Handler: Error: No SPIFFE ID
            Handler->>Handler: Reject request (401)
        else SPIFFE ID found
            SPIFFE-->>Handler: spiffeID string

            Handler->>Validation: ValidateSPIFFEID(spiffeID)

            Validation->>Validation: Check format
            Note right of Validation: Must start with "spiffe://"<br/>Valid trust domain<br/>Valid path

            alt Invalid format
                Validation-->>Handler: Error: Invalid SPIFFE ID
                Handler->>Handler: Reject request (401)
            else Valid format
                Validation-->>Handler: SPIFFE ID validated

                Handler->>Handler: Use SPIFFE ID for authorization
                Note right of Handler: Check policies<br/>Determine permissions

                Handler->>Handler: Process request
            end
        end
    end
```

**Key Functions:**

```go
func extractSPIFFEID(r *http.Request) (string, error) {
    if r.TLS == nil {
        return "", errors.New("no TLS connection")
    }

    if len(r.TLS.PeerCertificates) == 0 {
        return "", errors.New("no client certificate")
    }

    cert := r.TLS.PeerCertificates[0]

    // Extract SPIFFE ID from SAN
    for _, uri := range cert.URIs {
        if strings.HasPrefix(uri.String(), "spiffe://") {
            return uri.String(), nil
        }
    }

    return "", errors.New("no SPIFFE ID in certificate")
}

func validateSPIFFEID(spiffeID string) error {
    if !strings.HasPrefix(spiffeID, "spiffe://") {
        return errors.New("invalid SPIFFE ID format")
    }

    // Additional validation
    // - Trust domain matches expected
    // - Path is well-formed

    return nil
}
```

**Key Files:**
- `internal/auth/spiffe.go::IDFromRequest()`
- `internal/validation/spiffeid.go::ValidateSPIFFEID()`

---

## 5. mTLS with Peer Validation (Predicate)

Some SPIKE components restrict which peers they accept connections from.

```mermaid
sequenceDiagram
    participant Keeper as SPIKE Keeper
    participant Predicate as Peer Predicate
    participant Source as X509Source
    participant TLS as TLS Listener
    participant Client as Client<br/>(Nexus or Bootstrap)

    Note over Keeper: SPIKE Keeper starts

    Keeper->>Predicate: Define allowed peers
    Note right of Predicate: AllowKeeperPeer:<br/>- spike/nexus<br/>- spike/bootstrap

    Keeper->>Source: Get X509Source

    Keeper->>TLS: Create TLS listener with predicate

    TLS->>TLS: Configure TLS
    Note right of TLS: tlsConfig := &tls.Config{<br/>  GetCertificate: source.GetCertificate,<br/>  GetClientCAs: source.GetClientCAs,<br/>  ClientAuth: tls.RequireAndVerifyClientCert,<br/>  VerifyPeerCertificate: predicate.Verify,<br/>}

    Note over TLS: Server listening on port 8554

    Client->>TLS: Connect (TLS handshake)

    TLS->>TLS: Verify client certificate signature
    Note right of TLS: Standard TLS verification

    TLS->>Predicate: VerifyPeerCertificate(rawCerts, chains)

    Predicate->>Predicate: Extract SPIFFE ID from cert

    Predicate->>Predicate: Check against allowed list
    Note right of Predicate: Is peer spike/nexus or<br/>spike/bootstrap?

    alt Peer not allowed
        Predicate-->>TLS: Error: Peer not authorized
        TLS-->>Client: TLS handshake failed (bad certificate)
        Note over Client: Connection refused
    else Peer allowed
        Predicate-->>TLS: Peer authorized

        TLS-->>Client: TLS handshake complete

        Note over Client,TLS: mTLS connection established

        Client->>Keeper: Send HTTP request

        Keeper->>Keeper: Process request

        Keeper-->>Client: HTTP response
    end
```

**Predicate Example:**

```go
var AllowKeeperPeer = func(id spiffeid.ID) error {
    idStr := id.String()

    // Allow SPIKE Nexus
    if spiffeid.IsNexus(idStr) {
        return nil
    }

    // Allow SPIKE Bootstrap
    if spiffeid.IsBootstrap(idStr) {
        return nil
    }

    // Reject all others
    return fmt.Errorf("peer %s not authorized", idStr)
}
```

**Key Files:**
- `internal/auth/predicate/predicate.go` - Peer predicates
- `app/keeper/internal/net/serve.go::ServeWithPredicate()`

---

## 6. Complete mTLS Handshake Flow

```mermaid
sequenceDiagram
    participant Client as Client<br/>(with SVID A)
    participant ClientTLS as Client TLS Stack
    participant ServerTLS as Server TLS Stack
    participant Server as Server<br/>(with SVID B)
    participant CA as Trust Bundle<br/>(CA Certs)

    Note over Client,Server: Establish mTLS connection

    Client->>ClientTLS: Connect to server

    ClientTLS->>ServerTLS: ClientHello
    Note right of ClientTLS: Supported ciphers,<br/>TLS versions

    ServerTLS->>Server: Get server certificate
    Server->>Server: source.GetX509SVID()
    Server-->>ServerTLS: SVID B (cert + key)

    ServerTLS->>ClientTLS: ServerHello + Certificate B
    Note right of ServerTLS: Server's SPIFFE SVID

    ServerTLS->>ClientTLS: CertificateRequest
    Note right of ServerTLS: Require client certificate

    ClientTLS->>Client: Get client certificate
    Client->>Client: source.GetX509SVID()
    Client-->>ClientTLS: SVID A (cert + key)

    ClientTLS->>ServerTLS: Certificate A
    Note right of ClientTLS: Client's SPIFFE SVID

    par Verify server certificate (client side)
        ClientTLS->>CA: Get trust bundle
        CA-->>ClientTLS: CA certificates

        ClientTLS->>ClientTLS: Verify Certificate B signature
        Note right of ClientTLS: Check B signed by CA

        ClientTLS->>ClientTLS: Extract SPIFFE ID from B
        Note right of ClientTLS: Optionally verify<br/>expected SPIFFE ID

        alt Verification failed
            ClientTLS->>Client: Abort handshake
        end
    and Verify client certificate (server side)
        ServerTLS->>CA: Get trust bundle
        CA-->>ServerTLS: CA certificates

        ServerTLS->>ServerTLS: Verify Certificate A signature
        Note right of ServerTLS: Check A signed by CA

        ServerTLS->>Server: VerifyPeerCertificate(A)
        Note right of Server: Custom predicate check<br/>Extract SPIFFE ID

        alt Verification failed
            Server->>ServerTLS: Reject peer
            ServerTLS->>ClientTLS: Abort handshake
        end
    end

    ClientTLS->>ServerTLS: Finished
    ServerTLS->>ClientTLS: Finished

    Note over ClientTLS,ServerTLS: mTLS established.<br/>Encrypted channel ready.

    Client->>Server: Application data (HTTP request)
    Server->>Client: Application data (HTTP response)
```

**Security Properties:**
- **Mutual authentication**: Both client and server prove identity
- **Confidentiality**: All data encrypted (TLS 1.3 AES-GCM)
- **Integrity**: Tampering detected (TLS MAC)
- **Forward secrecy**: Ephemeral keys (ECDHE)
- **Certificate validation**: Both sides verify signatures
- **Identity binding**: SPIFFE IDs bound to certificates

---

## 7. SVID Rotation

```mermaid
sequenceDiagram
    participant Workload as SPIKE Component
    participant Source as X509Source
    participant Agent as SPIRE Agent
    participant TLS as Active TLS Connections

    Note over Workload: Component running with SVID

    Note over Source: SVID valid for ~1 hour<br/>Rotation before expiration

    Agent->>Source: Push updated SVID
    Note right of Agent: Proactive renewal<br/>~50% through TTL

    Source->>Source: Update internal state
    Note right of Source: New cert + private key<br/>Old SVID still valid

    Source->>Source: Notify watchers (if any)

    Note over TLS: Existing connections unaffected.<br/>Continue using old SVID.

    loop New connections
        Workload->>Source: GetX509SVID()
        Source-->>Workload: New SVID

        Workload->>TLS: Establish new connection
        Note right of TLS: Uses new SVID

        Note over TLS: New connection uses new SVID.<br/>Seamless rotation.
    end

    Note over Source: Old SVID expires.<br/>No longer valid.

    Note over TLS: All connections now use new SVID.<br/>Zero downtime rotation.
```

**Rotation Characteristics:**
- **Automatic**: No manual intervention
- **Proactive**: Renewed before expiration
- **Seamless**: Zero downtime
- **Concurrent**: Old and new SVIDs valid during rotation
- **Transparent**: Application code unaffected

---

## 8. Trust Domain and Federation

**Single Trust Domain (Common):**
```
Trust Domain: example.org
CA: SPIRE Server CA

All components trust same CA.
All SPIFFE IDs: spiffe://example.org/...
```

**Federated Trust Domains (Advanced):**
```
Trust Domain A: company-a.com
Trust Domain B: company-b.com

Components in A can communicate with B if:
- Trust bundles federated
- Both sides accept peer's trust domain
- Policies allow cross-domain access
```

**SPIKE Federation Support:**
- Currently: Single trust domain
- Future: Cross-trust-domain policies
- Use case: Multi-organization secrets sharing

---

## 9. Key Files Reference

**SPIFFE Integration:**
- `internal/net/spiffe.go::Source()` - Acquire X509Source
- `internal/auth/spiffe.go::IDFromRequest()` - Extract SPIFFE ID from request
- `internal/validation/spiffeid.go::ValidateSPIFFEID()` - Validate SPIFFE ID format

**mTLS Server:**
- `app/nexus/internal/net/serve.go::Serve()` - Start Nexus mTLS server
- `app/keeper/internal/net/serve.go::ServeWithPredicate()` - Start Keeper with predicate

**mTLS Client:**
- `internal/net/client.go::CreateMTLSClient()` - Create basic mTLS client
- `internal/net/client.go::CreateMTLSClientWithPredicate()` - Create client with server validation

**Predicates:**
- `internal/auth/predicate/predicate.go` - Peer validation predicates

**SPIFFE ID Checks:**
- `internal/auth/spiffeid/checks.go` - IsNexus, IsKeeper, IsSuperuser, etc.

---

## 10. Configuration

**Environment Variables:**

```bash
# SPIFFE Workload API socket
export SPIFFE_ENDPOINT_SOCKET=unix:///run/spire/sockets/agent.sock

# Trust domain (usually inferred from SVID)
export SPIKE_TRUST_DOMAIN=example.org

# Server ports
export SPIKE_NEXUS_TLS_PORT=8553
export SPIKE_KEEPER_TLS_PORT=8554
```

**SPIRE Configuration (example):**

```hcl
# SPIRE Agent configuration
agent {
    data_dir = "/opt/spire/data/agent"
    server_address = "spire-server"
    server_port = 8081
    trust_domain = "example.org"
}

# Workload API configuration
plugins {
    WorkloadAttestor "unix" {
        plugin_data {}
    }

    WorkloadAttestor "k8s" {
        plugin_data {
            kubelet_read_only_port = "10255"
        }
    }
}
```

**Registration Entry (example):**

```bash
# Register SPIKE Nexus
spire-server entry create \
  -spiffeID spiffe://example.org/spike/nexus \
  -parentID spiffe://example.org/spire/agent/k8s_psat/node1 \
  -selector k8s:ns:spike \
  -selector k8s:pod-label:app:spike-nexus

# Register SPIKE Keeper
spire-server entry create \
  -spiffeID spiffe://example.org/spike/keeper/0 \
  -parentID spiffe://example.org/spire/agent/k8s_psat/node2 \
  -selector k8s:ns:spike \
  -selector k8s:pod-label:app:spike-keeper \
  -selector k8s:pod-label:keeper-id:0
```

---

## Summary

**mTLS with SPIFFE:**
- **SVID**: X.509 certificate with SPIFFE ID in SAN
- **Workload API**: Automatic SVID acquisition from SPIRE Agent
- **Rotation**: Automatic certificate renewal (zero downtime)
- **Mutual Authentication**: Both client and server verify identity
- **Trust Model**: All workloads trust SPIRE CA
- **Zero Configuration**: No manual certificate management

**Key Properties:**
- **Identity-based**: Authorization based on SPIFFE ID (not IP, hostname)
- **Dynamic**: SVIDs issued at runtime (no pre-provisioned certs)
- **Short-lived**: SVIDs expire quickly (default: 1 hour)
- **Automatic Rotation**: Seamless renewal before expiration
- **Scalable**: Works for thousands of workloads

**Security Benefits:**
- **No shared secrets**: Each workload has unique SVID
- **No password management**: Certificate-based authentication
- **Defense in depth**: Multiple layers (TLS + SPIFFE ID validation)
- **Principle of least privilege**: Fine-grained policies per SPIFFE ID
- **Auditability**: All connections tied to workload identity
