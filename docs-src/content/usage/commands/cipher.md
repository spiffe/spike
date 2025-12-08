+++
#    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
#  \\\\\ Copyright 2024-present SPIKE contributors.
# \\\\\\\ SPDX-License-Identifier: Apache-2.

title = "spike cipher"
weight = 5
sort_by = "weight"
+++

# `spike cipher`

The `spike cipher` command provides **encryption and decryption** capabilities
through **SPIKE Nexus**. It allows workloads to protect sensitive data in
transit or at rest using cryptographic operations managed by the secrets
infrastructure.

## Quick Start

```bash
# Encrypt a file
spike cipher encrypt --file secret.txt --out secret.enc

# Decrypt a file
spike cipher decrypt --file secret.enc --out secret.txt

# Stream encryption/decryption via stdin/stdout
echo "sensitive data" | spike cipher encrypt | spike cipher decrypt
```

## What is SPIKE Cipher?

The cipher commands provide a secure way to encrypt and decrypt data using keys
managed by **SPIKE Nexus**. This enables:

* **Data protection**: Encrypt sensitive files or data streams
* **Key management**: Cryptographic keys are managed centrally by SPIKE Nexus
* **Access control**: Encryption operations are subject to SPIFFE-based
  authentication
* **Flexibility**: Support for both file-based and streaming operations

## Commands

### `spike cipher encrypt`

```bash
spike cipher encrypt [--file=<input>] [--out=<output>]
```

Encrypts data via **SPIKE Nexus**. The command supports two modes of operation:

#### Stream Mode (default)

Reads data from a file or stdin and writes encrypted data to a file or stdout.
This mode handles binary data transparently.

#### JSON Mode

When `--plaintext` is provided, the command accepts base64-encoded plaintext
and returns a JSON-formatted encryption result.

#### Flags:

| Flag           | Description                                          |
|----------------|------------------------------------------------------|
| `--file`, `-f` | Input file path (default: stdin)                     |
| `--out`, `-o`  | Output file path (default: stdout)                   |
| `--plaintext`  | Base64-encoded plaintext for JSON mode               |
| `--algorithm`  | Algorithm hint for JSON mode                         |

#### Examples:

```bash
# Encrypt a file to another file
spike cipher encrypt --file secret.txt --out secret.enc

# Encrypt from stdin to stdout
cat secret.txt | spike cipher encrypt > secret.enc

# Encrypt using short flags
spike cipher encrypt -f secret.txt -o secret.enc

# Encrypt with JSON mode (base64 input)
spike cipher encrypt --plaintext "c2Vuc2l0aXZlIGRhdGE="
```

### `spike cipher decrypt`

```bash
spike cipher decrypt [--file=<input>] [--out=<output>]
```

Decrypts data via **SPIKE Nexus**. The command supports two modes of operation:

#### Stream Mode (default)

Reads encrypted data from a file or stdin and writes decrypted plaintext to a
file or stdout. This mode handles binary data transparently.

#### JSON Mode

When `--version`, `--nonce`, or `--ciphertext` is provided, the command accepts
base64-encoded encryption components and returns plaintext output.

#### Flags:

| Flag           | Description                                          |
|----------------|------------------------------------------------------|
| `--file`, `-f` | Input file path (default: stdin)                     |
| `--out`, `-o`  | Output file path (default: stdout)                   |
| `--version`    | Version byte (0-255) for JSON mode                   |
| `--nonce`      | Base64-encoded nonce for JSON mode                   |
| `--ciphertext` | Base64-encoded ciphertext for JSON mode              |
| `--algorithm`  | Algorithm hint for JSON mode                         |

#### Examples:

```bash
# Decrypt a file to another file
spike cipher decrypt --file secret.enc --out secret.txt

# Decrypt from stdin to stdout
cat secret.enc | spike cipher decrypt > secret.txt

# Decrypt using short flags
spike cipher decrypt -f secret.enc -o secret.txt

# Decrypt with JSON mode components
spike cipher decrypt --version=1 --nonce="..." --ciphertext="..."
```

## Use Cases

### Encrypting Configuration Files

```bash
# Encrypt a configuration file before storing
spike cipher encrypt -f config.yaml -o config.yaml.enc

# Decrypt when needed
spike cipher decrypt -f config.yaml.enc -o config.yaml
```

### Pipeline Processing

```bash
# Process data through encryption in a pipeline
generate-secrets | spike cipher encrypt | store-encrypted-data

# Decrypt and process
fetch-encrypted-data | spike cipher decrypt | process-secrets
```

### Backup Encryption

```bash
# Encrypt a database dump
pg_dump mydb | spike cipher encrypt > backup.enc

# Restore from encrypted backup
spike cipher decrypt -f backup.enc | psql mydb
```

## Best Practices

* Use file-based operations for large data to avoid memory issues
* Pipe operations are useful for automation and scripting
* Ensure the workload has appropriate SPIFFE credentials before encryption
* Store encrypted files securely; encryption adds a layer but is not a
  replacement for access control
* Use consistent encryption for data that will be decrypted later

## Technical Details

### Cryptographic Algorithm

**SPIKE Cipher** uses **AES-256-GCM** (Galois/Counter Mode) for authenticated
encryption:

| Property            | Value                                             |
|---------------------|---------------------------------------------------|
| Algorithm           | AES-256-GCM                                       |
| Key Size            | 32 bytes (256 bits)                               |
| Nonce Size          | 12 bytes (96 bits)                                |
| Authentication      | Built-in (AEAD)                                   |

**Security Properties:**

* **Confidentiality**: Plaintext is hidden from unauthorized parties
* **Integrity**: Any tampering or corruption is detected
* **Authenticity**: Verifies data originated from a valid source
* **Freshness**: Unique nonce prevents replay attacks

### Stream Mode Binary Format

In stream mode, the encrypted output has the following binary format:

```text
+--------+-------------+------------------+
| Byte 0 | Bytes 1-12  | Bytes 13 to end  |
+--------+-------------+------------------+
| 0x01   | Nonce (GCM) | Ciphertext       |
+--------+-------------+------------------+
```

* **Byte 0**: Version byte (currently `0x01`)
* **Bytes 1-12**: 12-byte GCM nonce (randomly generated)
* **Bytes 13+**: The actual ciphertext with authentication tag

### JSON vs Stream Mode

| Aspect           | JSON Mode              | Stream Mode                |
|------------------|------------------------|----------------------------|
| Content-Type     | `application/json`     | `application/octet-stream` |
| Request Format   | JSON object            | Binary bytes               |
| Response Format  | JSON object            | Binary bytes               |
| Overhead         | Higher (JSON encoding) | Lower (raw binary)         |
| Use Case         | Small data, structured | Large data, efficiency     |
| Nonce Location   | JSON field             | Bytes 1-12 of response     |
| Version Location | JSON field             | Byte 0 of response         |

## Security Considerations

* All cipher operations require valid SPIFFE authentication
* Encryption keys are managed by **SPIKE Nexus** and never exposed to clients
* The cipher operations use authenticated encryption (AEAD)
* Memory containing sensitive data is cleared after operations
* Nonces are randomly generated and must be unique per encryption

----

## `spike` Command Index

{{ toc_commands() }}

<p>&nbsp;</p>

----

{{ toc_usage() }}

----

{{ toc_top() }}