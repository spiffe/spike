![SPIKE](../assets/spike-banner-lg.png)

```mermaid
sequenceDiagram
    participant Client as Client Application
    participant Auth as Auth Layer
    participant Handler as Cipher Handler
    participant Guard as Guard Validation
    participant Backend as State Backend
    participant Crypto as AES-GCM Cipher

    Note over Client: Client wants to encrypt<br/>large binary data

    Client->>Handler: POST /v1/cipher/encrypt<br/>Content-Type: application/octet-stream<br/>[binary data]
    Note right of Client: mTLS with SVID<br/>Body: raw bytes

    Handler->>Auth: Extract SPIFFE ID from cert
    Auth-->>Handler: spiffeID string

    Handler->>Handler: handleStreamingEncrypt(r, spiffeID)

    Handler->>Handler: Read binary stream from body
    Note right of Handler: ioutil.ReadAll(r.Body)

    Handler->>Handler: Construct CipherEncryptRequest
    Note right of Handler: plaintext = binary data<br/>algorithm = "" (default AES)

    Handler->>Guard: guardEncryptRequest(spiffeID, req)
    Guard->>Guard: Validate plaintext not empty
    Guard-->>Handler: Validation passed

    Handler->>Backend: GetCipher()
    Backend-->>Handler: cipher.AEAD (AES-GCM)

    Handler->>Crypto: Generate 12-byte nonce
    Crypto-->>Handler: nonce []byte

    Handler->>Crypto: cipher.Seal(nil, nonce, plaintext, nil)
    Crypto-->>Handler: ciphertext []byte

    Handler->>Handler: Build binary response
    Note right of Handler: Format:<br/>[version:1byte][nonce:12bytes][ciphertext:variable]

    Handler->>Handler: Concatenate: version + nonce + ciphertext

    Handler-->>Client: 200 OK<br/>Content-Type: application/octet-stream<br/>[1 byte version][12 bytes nonce][N bytes ciphertext]

    Note over Client: Parse binary response:<br/>byte 0: version<br/>bytes 1-12: nonce<br/>bytes 13+: ciphertext
```
