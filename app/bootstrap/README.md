![SPIKE](../../assets/spike-banner-lg.png)

# SPIKE Bootstrap

**SPIKE Bootstrap** is a critical initialization service that generates and distributes cryptographic root key shares to SPIKE Keeper instances using 
Shamir's Secret Sharing scheme. It is responsible for establishing the 
foundational cryptographic trust infrastructure for the SPIKE system.

## Overview

The bootstrap module performs a one-time initialization process that:

* Generates a cryptographically secure random root key
* Splits the root key into multiple shares using Shamir's Secret Sharing
* Distributes each share to a corresponding SPIKE Keeper instance
* Establishes the initial trust foundation for the entire SPIKE system

## Key Components

### Secret Share Generation (`internal/state/`)

* **RootShares()**: Generates a 32-byte random root key and splits it into `n` shares using Shamir's Secret Sharing with threshold `t`
* **KeeperShare()**: Retrieves the specific share corresponding to a given Keeper ID
* Uses P256 elliptic curve group for cryptographic operations
* Implements deterministic share generation for crash recovery consistency

### Environment Configuration (`internal/env/`)

* **ShamirShares()**: Configures total number of shares (default: 3)
* **ShamirThreshold()**: Sets minimum shares needed for reconstruction (default: 2)
* **Keepers()**: Parses `SPIKE_NEXUS_KEEPER_PEERS` environment variable for Keeper endpoints
* **TrustRootForKeeper()**: Configures trusted domain validation for Keeper connections

### Network Communication (`internal/net/`)

* **Source()**: Establishes SPIFFE workload API connection for X.509 credentials
* **MTLSClient()**: Creates mTLS HTTP client with peer validation against trusted Keeper identities
* **Post()**: Sends secret shares to Keeper instances via HTTP POST requests
* **Payload()**: Marshals share data into `ShardContributionRequest` format

### Security Validation (`internal/validation/`)

* **SanityCheck()**: Verifies share reconstruction works correctly before distribution
* Ensures cryptographic integrity of the generated shares
* Performs memory cleanup of sensitive data

### URL Construction (`internal/url/`)
* **KeeperEndpoint()**: Builds complete API endpoints for Keeper contribution requests

## Usage

The bootstrap service is executed once during system initialization:

```bash
bootstrap -init
```

### Environment Variables

* **SPIKE_NEXUS_SHAMIR_SHARES**: Total number of shares (default: 3)
* **SPIKE_NEXUS_SHAMIR_THRESHOLD**: Minimum shares for reconstruction (default: 2)
* **SPIKE_NEXUS_KEEPER_PEERS**: Comma-separated list of Keeper HTTPS URLs
* **SPIKE_TRUST_ROOT_KEEPER**: Trust root domain for Keeper validation (default: "spike.ist")

### Example Configuration

```bash
export SPIKE_NEXUS_SHAMIR_SHARES=5
export SPIKE_NEXUS_SHAMIR_THRESHOLD=3
export SPIKE_NEXUS_KEEPER_PEERS="https://keeper1:8443,https://keeper2:8543,https://keeper3:8643"
export SPIKE_TRUST_ROOT_KEEPER="spike.ist"
```

## Security Features

* **Cryptographic Randomness**: Uses cryptographically secure random number generation
* **mTLS Authentication**: All communication with Keepers uses mutual TLS with SPIFFE identity validation
* **Secret Sharing**: Implements Shamir's Secret Sharing for distributed trust
* **Memory Protection**: Automatically zeroes sensitive data after use
* **Peer Validation**: Only accepts connections from validated Keeper identities
* **HTTPS Enforcement**: Requires all Keeper endpoints to use HTTPS

## Architecture

The bootstrap process follows this sequence:

1. **Initialization**: Parse command line flags and environment configuration
2. **Key Generation**: Create cryptographically secure random root key
3. **Share Creation**: Split root key into `n` shares using Shamir's Secret Sharing
4. **Validation**: Verify shares can reconstruct the original secret
5. **Distribution**: Send each share to its corresponding Keeper instance
6. **Cleanup**: Securely dispose of sensitive data

## Dependencies

* **SPIFFE Workload API**: For X.509 credential management
* **Cloudflare CIRCL**: For cryptographic operations and Shamir's Secret Sharing
* **SPIKE SDK**: For SPIFFE integration and networking utilities

## Error Handling

The service terminates with exit code 1 if:
* Required environment variables are missing
* Cryptographic operations fail
* Network communication errors occur
* Share validation fails
* Invalid Keeper configurations are detected
