![SPIKE](../../assets/spike-banner-lg.png)

## SPIKE Nexus

**SPIKE Nexus** is the secrets store of **SPIKE**. It provides encrypted secret
storage with versioning, soft-delete support, and SPIFFE-based access control.

### Storage Backends

SPIKE Nexus supports multiple storage backends, configured via the
`SPIKE_NEXUS_BACKEND_STORE` environment variable:

* **sqlite** (production): Persistent encrypted storage using SQLite. Secrets
  are stored in `~/.spike/data/spike.db` with AES-256-GCM encryption at rest.
* **memory** (development/testing): In-memory storage. Data is lost on restart.
* **lite**: Encryption-only mode for encryption-as-a-service use cases. No
  secret persistence; provides cipher access for external encryption needs.

### Root Key Management

Secrets are encrypted using a **root key** that is automatically generated
during **SPIKE Nexus**'s bootstrapping sequence. This root key is securely
shared with **SPIKE Keeper** instances for redundancy and automatic recovery.

The administrator installing **SPIKE** for the first time is encouraged to take
an encrypted backup of the root key to enable manual recovery if the system
becomes locked.

When initializing **SPIKE** from the **SPIKE Pilot** CLI, you will be prompted
to enter a password to back up the root key. It is **crucial** that you do not
forget the password and do not lose the encrypted backup of the root key.
