![SPIKE](../../assets/spike-banner-lg.png)

## SPIKE Bootstrap

**SPIKE Bootstrap** is a critical initialization service that generates and 
distributes cryptographic root key shards to **SPIKE Keeper** instances using 
[Shamir's Secret Sharing scheme][shamir]. It is responsible for establishing the 
foundational cryptographic trust infrastructure for the SPIKE system.

[shamir]: https://en.wikipedia.org/wiki/Shamir%27s_Secret_Sharing "Shamir's Secret Sharing"

## Overview

The bootstrap module performs a one-time initialization process that:

* Generates a cryptographically secure random root key
* Splits the root key into multiple shares using Shamir's Secret Sharing
* Distributes each share to a corresponding SPIKE Keeper instance
* Establishes the initial trust foundation for the entire SPIKE system

## Configuration

Boostrap reads environment variables to configure its behavior:

* `SPIKE_NEXUS_API_URL` (*default: `https://localhost:8553`*)
* `SPIFFE_ENDPOINT_SOCKET` (*default: 
  `unix:///spiffe-workload-api/spire-agent.sock`*)
* `SPIKE_TRUST_ROOT` (*default `spike.ist`*)
* `SPIKE_NEXUS_SHAMIR_SHARES` (*default: `3`*)
* `SPIKE_NEXUS_SHAMIR_THRESHOLD` (*default: `2`*)
* `SPIKE_NEXUS_KEEPER_PEERS` (*comma-delimited list of SPIKE Keeper HTTPS URLs 
  to seed the trust foundation*)

## Usage

The bootstrap service is executed once during system initialization:

```bash
bootstrap -init
```

## Security Features

* **Cryptographic Randomness**: Uses cryptographically secure random seed 
  generation
* **mTLS Authentication**: All communication with Keepers uses mutual TLS with 
  **SPIFFE** identity validation
* **Secret Sharing**: Implements Shamir's Secret Sharing for distributed trust
* **Memory Protection**: Automatically zeroes sensitive data after use
* **Peer Validation**: Only accepts connections from validated **SPIKE Keeper** 
  identities

## Architecture

The bootstrap process follows this sequence:

1. **Initialization**: Parse command line flags and environment configuration
2. **Key Generation**: Create a cryptographically secure random root key
3. **Share Creation**: Split the root key into `n` shares using Shamir's Secret 
   Sharing
4. **Validation**: Verify shares can reconstruct the original secret
5. **Distribution**: Send each share to its corresponding Keeper instance
6. **Cleanup**: Securely dispose of sensitive data

## Dependencies

* **SPIFFE Workload API**: For X.509 credential management
* **Cloudflare CIRCL**: For cryptographic operations and Shamir's Secret Sharing
* **SPIKE SDK**: For SPIFFE integration and networking utilities

## Error Handling

The service terminates with exit code `1` if:

* Required environment variables are missing
* Cryptographic operations fail
* Network communication errors occur
* Share validation fails
* Invalid Keeper configurations are detected
