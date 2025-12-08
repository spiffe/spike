![SPIKE](../assets/spike-banner-lg.png)

```mermaid
sequenceDiagram
participant Client as Client Application
participant Auth as Auth Layer
participant Handler as Cipher Handler
participant Guard as Guard Validation
participant Backend as State Backend
participant Crypto as AES-GCM Cipher

    Note over Client: Client wants to encrypt data

    Client->>Handler: POST /v1/cipher/encrypt<br/>Content-Type: application/json<br/>{plaintext: []byte, algorithm: "aes"}
    Note right of Client: mTLS with SVID

    Handler->>Auth: Extract SPIFFE ID from cert
    Auth->>Auth: Validate SPIFFE ID format
    Auth-->>Handler: spiffeID string

    Handler->>Handler: handleJSONEncrypt(r, spiffeID)

    Handler->>Handler: Parse JSON request body
    Note right of Handler: Unmarshal to CipherEncryptRequest

    Handler->>Guard: guardEncryptRequest(spiffeID, req)

    Guard->>Guard: Validate plaintext not empty
    Guard->>Guard: Validate algorithm (if specified)
    Guard->>Guard: Check permissions
    Note right of Guard: Future: Check if client<br/>can encrypt data

    Guard-->>Handler: Validation passed

    Handler->>Backend: GetCipher()
    Note right of Backend: Retrieve cipher created<br/>from root key

    Backend-->>Handler: cipher.AEAD (AES-GCM)

    Handler->>Crypto: Generate 12-byte nonce
    Note right of Crypto: crypto/rand.Read()
    Crypto-->>Handler: nonce []byte

    Handler->>Crypto: cipher.Seal(nil, nonce, plaintext, nil)
    Note right of Crypto: AES-256-GCM encryption<br/>Authenticated encryption
    Crypto-->>Handler: ciphertext []byte

    Handler->>Handler: Build CipherEncryptResponse
    Note right of Handler: version: 1<br/>nonce: []byte<br/>ciphertext: []byte

    Handler-->>Client: 200 OK<br/>{version: 1, nonce: [12]byte,<br/>ciphertext: []byte}

    Note over Client: Store nonce and ciphertext<br/>for later decryption
```