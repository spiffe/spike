![SPIKE](../assets/spike-banner-lg.png)

```mermaid
sequenceDiagram
    participant Client as Client Application
    participant Auth as Auth Layer
    participant Handler as Cipher Handler
    participant Guard as Guard Validation
    participant Backend as State Backend
    participant Crypto as AES-GCM Cipher

    Note over Client: Client wants to decrypt<br/>binary ciphertext

    Client->>Handler: POST /v1/cipher/decrypt<br/>Content-Type: application/octet-stream<br/>[version][nonce][ciphertext]
    Note right of Client: mTLS with SVID<br/>Binary format from encrypt

    Handler->>Auth: Extract SPIFFE ID from cert
    Auth-->>Handler: spiffeID string

    Handler->>Handler: handleStreamingDecrypt(r, spiffeID)

    Handler->>Handler: Read binary stream from body

    Handler->>Handler: Parse binary format
    Note right of Handler: byte 0: version<br/>bytes 1-12: nonce<br/>bytes 13+: ciphertext

    Handler->>Handler: Extract components
    Note right of Handler: version = data[0]<br/>nonce = data[1:13]<br/>ciphertext = data[13:]

    Handler->>Handler: Construct CipherDecryptRequest

    Handler->>Guard: guardDecryptRequest(spiffeID, req)
    Guard->>Guard: Validate version is 1
    Guard->>Guard: Validate nonce length is 12
    Guard->>Guard: Validate ciphertext not empty
    Guard-->>Handler: Validation passed

    Handler->>Backend: GetCipher()
    Backend-->>Handler: cipher.AEAD (AES-GCM)

    Handler->>Crypto: cipher.Open(nil, nonce, ciphertext, nil)

    alt Decryption successful
        Crypto-->>Handler: plaintext []byte

        Handler-->>Client: 200 OK<br/>Content-Type: application/octet-stream<br/>[plaintext binary data]
    else Authentication failed
        Crypto-->>Handler: error (tag mismatch)

        Handler-->>Client: 401 Unauthorized

        Note over Client: Data tampered or wrong key
    end
```
