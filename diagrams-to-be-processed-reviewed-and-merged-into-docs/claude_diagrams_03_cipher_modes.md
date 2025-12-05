# SPIKE Cipher Operations (JSON and Stream Modes)

## Overview

SPIKE Nexus provides encryption and decryption services through two modes:
1. **JSON Mode**: Structured requests/responses with JSON encoding
2. **Stream Mode**: Binary data for efficiency with large payloads

Both modes use AES-256-GCM for authenticated encryption.

---

## 1. JSON Mode Encryption

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

---

## 2. JSON Mode Decryption

```mermaid
sequenceDiagram
    participant Client as Client Application
    participant Auth as Auth Layer
    participant Handler as Cipher Handler
    participant Guard as Guard Validation
    participant Backend as State Backend
    participant Crypto as AES-GCM Cipher

    Note over Client: Client wants to decrypt data

    Client->>Handler: POST /v1/cipher/decrypt<br/>Content-Type: application/json<br/>{version: 1, nonce: []byte,<br/>ciphertext: []byte}
    Note right of Client: mTLS with SVID

    Handler->>Auth: Extract SPIFFE ID from cert
    Auth->>Auth: Validate SPIFFE ID format
    Auth-->>Handler: spiffeID string

    Handler->>Handler: handleJSONDecrypt(r, spiffeID)

    Handler->>Handler: Parse JSON request body
    Note right of Handler: Unmarshal to CipherDecryptRequest

    Handler->>Guard: guardDecryptRequest(spiffeID, req)

    Guard->>Guard: Validate version is 1
    Guard->>Guard: Validate nonce length (12 bytes)
    Guard->>Guard: Validate ciphertext not empty
    Guard->>Guard: Check permissions
    Note right of Guard: Future: Check if client<br/>can decrypt data

    Guard-->>Handler: Validation passed

    Handler->>Backend: GetCipher()
    Backend-->>Handler: cipher.AEAD (AES-GCM)

    Handler->>Crypto: cipher.Open(nil, nonce, ciphertext, nil)
    Note right of Crypto: AES-256-GCM decryption<br/>Verifies authentication tag

    alt Decryption successful
        Crypto-->>Handler: plaintext []byte

        Handler->>Handler: Build CipherDecryptResponse
        Note right of Handler: plaintext: []byte

        Handler-->>Client: 200 OK<br/>{plaintext: []byte}
    else Authentication failed
        Crypto-->>Handler: error (tag mismatch)

        Handler-->>Client: 401 Unauthorized<br/>{err: "decryption failed"}

        Note over Client: Data tampered or wrong key
    end
```

---

## 3. Stream Mode Encryption

Binary mode for efficient handling of large data.

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

**Binary Response Format:**
```
+--------+-------------+------------------+
| Byte 0 | Bytes 1-12  | Bytes 13 to end  |
+--------+-------------+------------------+
| 0x01   | Nonce (GCM) | Ciphertext       |
+--------+-------------+------------------+
```

---

## 4. Stream Mode Decryption

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

---

## Key Differences Between Modes

| Aspect | JSON Mode | Stream Mode |
|--------|-----------|-------------|
| Content-Type | `application/json` | `application/octet-stream` |
| Request Format | JSON object | Binary bytes |
| Response Format | JSON object | Binary bytes |
| Overhead | Higher (JSON encoding) | Lower (raw binary) |
| Use Case | Small data, structured | Large data, efficiency |
| Nonce Location | JSON field | Bytes 1-12 of response |
| Version Location | JSON field | Byte 0 of response |

---

## Key Files

- `app/nexus/internal/route/cipher/encrypt.go` - Encryption handler
- `app/nexus/internal/route/cipher/decrypt.go` - Decryption handler
- `app/nexus/internal/route/cipher/handle.go` - Mode detection and routing
- `app/nexus/internal/route/cipher/crypto.go` - Crypto operations
- `internal/crypto/gcm.go` - GCM constants (nonce size, etc.)

---

## Cryptographic Details

**Algorithm:** AES-256-GCM (Galois/Counter Mode)

**Key Size:** 32 bytes (256 bits)
- Derived from root key
- Root key is 32-byte random value

**Nonce Size:** 12 bytes (96 bits)
- GCM standard size
- Randomly generated for each encryption
- MUST be unique for each encryption with same key

**Authentication Tag:** Automatically included in ciphertext
- GCM provides authenticated encryption (AEAD)
- Detects tampering or corruption
- No additional HMAC needed

**Security Properties:**
- **Confidentiality**: Plaintext hidden
- **Integrity**: Tampering detected
- **Authenticity**: Verifies data from correct source
- **Freshness**: Unique nonce prevents replay

---

## Configuration

Environment variables:
- `SPIKE_NEXUS_URL`: SPIKE Nexus endpoint
- `SPIKE_NEXUS_TLS_PORT`: mTLS port (default: 8553)

---

## Example Usage

### JSON Mode (Go SDK)
```go
client := api.NewClient(source)

resp, err := client.Encrypt([]byte("sensitive data"))
// resp.Nonce, resp.Ciphertext

plaintext, err := client.Decrypt(resp.Nonce, resp.Ciphertext)
```

### Stream Mode (cURL)
```bash
# Encrypt
echo "data" | curl -X POST \
  --data-binary @- \
  -H "Content-Type: application/octet-stream" \
  https://spike-nexus:8553/v1/cipher/encrypt > encrypted.bin

# Decrypt
curl -X POST \
  --data-binary @encrypted.bin \
  -H "Content-Type: application/octet-stream" \
  https://spike-nexus:8553/v1/cipher/decrypt
```
