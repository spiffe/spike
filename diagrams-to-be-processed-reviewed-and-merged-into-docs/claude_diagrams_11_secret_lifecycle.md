# Secret Lifecycle Management

## Overview

Secrets in SPIKE have a complete lifecycle from creation through versioning,
updates, soft deletion, and permanent removal. This document covers all
operations from CLI to database.

---

## 1. Secret Creation (Put)

```mermaid
sequenceDiagram
    participant User as Operator
    participant Pilot as SPIKE Pilot<br/>(spike secret put)
    participant Nexus as SPIKE Nexus
    participant Policy as Policy Engine
    participant State as State Manager
    participant Crypto as Encryption
    participant DB as SQLite Database

    Note over User: Create new secret

    User->>Pilot: spike secret put secrets/db/creds \<br/>username=admin password=secret123

    Pilot->>Pilot: Parse key=value pairs
    Note right of Pilot: Build map:<br/>{username: admin,<br/>password: secret123}

    Pilot->>Nexus: POST /v1/secret<br/>SecretPutRequest<br/>{path, values}
    Note right of Pilot: mTLS with SVID

    Nexus->>Nexus: Extract peer SPIFFE ID
    Note right of Nexus: From client certificate

    Nexus->>Policy: CheckPermission(spiffeID, path, write)

    Policy->>Policy: Load policies from cache
    Policy->>Policy: For each policy:<br/>- Match SPIFFE ID pattern?<br/>- Match path pattern?<br/>- Has write permission?

    alt No matching policy
        Policy-->>Nexus: Unauthorized
        Nexus-->>Pilot: 403 Forbidden
        Pilot-->>User: Error: Permission denied
    else Policy matched
        Policy-->>Nexus: Authorized

        Nexus->>State: UpsertSecret(path, values)

        State->>DB: SELECT * FROM secret_metadata<br/>WHERE path = ?
        Note right of DB: Check if secret exists

        alt Secret doesn't exist
            State->>State: Create new metadata
            Note right of State: current_version = 0<br/>oldest_version = 0<br/>max_versions = 10

            State->>DB: INSERT INTO secret_metadata<br/>(path, current_version,<br/>oldest_version, created_time,<br/>updated_time, max_versions)
        else Secret exists
            DB-->>State: Existing metadata
        end

        State->>State: Increment version
        Note right of State: newVersion = current_version + 1

        State->>State: Marshal values to JSON
        Note right of State: {"username": "admin",<br/>"password": "secret123"}

        State->>Crypto: encrypt(jsonData)

        Crypto->>Crypto: Generate 12-byte nonce
        Crypto->>Crypto: Encrypt with AES-256-GCM
        Crypto-->>State: nonce, ciphertext

        State->>DB: BEGIN TRANSACTION

        State->>DB: INSERT INTO secrets<br/>(path, version, nonce,<br/>encrypted_data, created_time)
        Note right of DB: path: secrets/db/creds<br/>version: 1<br/>nonce: [12 bytes]<br/>encrypted_data: [blob]<br/>created_time: timestamp

        State->>DB: UPDATE secret_metadata<br/>SET current_version = ?,<br/>updated_time = ?<br/>WHERE path = ?

        State->>State: Check version count
        Note right of State: If versions > max_versions,<br/>prune oldest

        alt Too many versions
            State->>DB: DELETE FROM secrets<br/>WHERE path = ? AND version < ?
            Note right of DB: Remove oldest versions

            State->>DB: UPDATE secret_metadata<br/>SET oldest_version = ?
        end

        State->>DB: COMMIT TRANSACTION

        State-->>Nexus: Secret stored (version N)

        Nexus-->>Pilot: 200 OK<br/>SecretPutResponse<br/>{version: N}

        Pilot-->>User: Secret created at version N
    end
```

**Key Files:**
- `app/spike/internal/cmd/secret/put.go` - CLI command
- `app/nexus/internal/route/secret/put.go` - HTTP handler
- `app/nexus/internal/state/base/secret.go::UpsertSecret()`
- `app/nexus/internal/state/backend/sqlite/persist/secret.go::StoreSecret()`

---

## 2. Secret Read (Get)

```mermaid
sequenceDiagram
    participant User as Application/Operator
    participant Client as SPIKE SDK/Pilot
    participant Nexus as SPIKE Nexus
    participant Policy as Policy Engine
    participant State as State Manager
    participant DB as SQLite Database
    participant Crypto as Decryption

    Note over User: Read secret

    User->>Client: Get secret: secrets/db/creds

    Client->>Nexus: POST /v1/secret/get<br/>SecretGetRequest<br/>{path, version (optional)}
    Note right of Client: mTLS with SVID

    Nexus->>Nexus: Extract peer SPIFFE ID

    Nexus->>Policy: CheckPermission(spiffeID, path, read)

    Policy->>Policy: Check policies

    alt No matching policy
        Policy-->>Nexus: Unauthorized
        Nexus-->>Client: 403 Forbidden
        Client-->>User: Error: Permission denied
    else Policy matched
        Policy-->>Nexus: Authorized

        Nexus->>State: GetSecret(path, version)

        alt Version not specified
            State->>DB: SELECT current_version<br/>FROM secret_metadata<br/>WHERE path = ?
            DB-->>State: current_version
        else Version specified
            Note over State: Use specified version
        end

        State->>DB: SELECT nonce, encrypted_data<br/>FROM secrets<br/>WHERE path = ? AND version = ?<br/>AND deleted_time IS NULL

        alt Secret not found
            DB-->>State: No rows
            State-->>Nexus: Error: Secret not found
            Nexus-->>Client: 404 Not Found
            Client-->>User: Error: Secret not found
        else Secret found
            DB-->>State: nonce, encrypted_data

            State->>Crypto: decrypt(nonce, ciphertext)

            Crypto->>Crypto: AES-256-GCM decryption
            Crypto-->>State: plaintext JSON

            State->>State: Unmarshal JSON
            Note right of State: {"username": "admin",<br/>"password": "secret123"}

            State-->>Nexus: Secret data (map)

            Nexus-->>Client: 200 OK<br/>SecretGetResponse<br/>{values, version}

            Client-->>User: {username: admin,<br/>password: secret123}
        end
    end
```

**Key Files:**
- `app/spike/internal/cmd/secret/get.go` - CLI command
- `app/nexus/internal/route/secret/get.go` - HTTP handler
- `app/nexus/internal/state/base/secret.go::GetSecret()`
- `app/nexus/internal/state/backend/sqlite/persist/secret.go::ReadSecret()`

---

## 3. Secret Update (New Version)

```mermaid
sequenceDiagram
    participant User as Operator
    participant Pilot as SPIKE Pilot
    participant Nexus as SPIKE Nexus
    participant State as State Manager
    participant DB as SQLite Database

    Note over User: Update existing secret

    User->>Pilot: spike secret put secrets/db/creds \<br/>username=admin password=newpass456

    Pilot->>Nexus: POST /v1/secret<br/>SecretPutRequest

    Nexus->>Nexus: Check authorization (write permission)

    Nexus->>State: UpsertSecret(path, newValues)

    State->>DB: SELECT current_version<br/>FROM secret_metadata<br/>WHERE path = ?
    DB-->>State: current_version = 1

    State->>State: newVersion = 2
    Note right of State: Increment version

    State->>State: Encrypt new data<br/>(generates new nonce)

    State->>DB: INSERT INTO secrets<br/>(path: secrets/db/creds,<br/>version: 2,<br/>nonce: [new 12 bytes],<br/>encrypted_data: [new blob],<br/>created_time: timestamp)

    State->>DB: UPDATE secret_metadata<br/>SET current_version = 2,<br/>updated_time = now<br/>WHERE path = secrets/db/creds

    Note over State,DB: Version 1 still exists in database<br/>Version 2 is now current

    State-->>Nexus: Secret updated (version 2)

    Nexus-->>Pilot: 200 OK<br/>{version: 2}

    Pilot-->>User: Secret updated to version 2
```

**Note:** Old versions are retained up to `max_versions` limit.

---

## 4. Secret Versioning

```mermaid
graph TD
    A[Secret: secrets/db/creds] --> B[Version 1<br/>created: 2024-01-01]
    A --> C[Version 2<br/>created: 2024-01-05]
    A --> D[Version 3<br/>created: 2024-01-10]
    A --> E[Version 4<br/>created: 2024-01-15]
    A --> F[Version 5 CURRENT<br/>created: 2024-01-20]

    style F fill:#4ecdc4
    style E fill:#95e1d3
    style D fill:#95e1d3
    style C fill:#95e1d3
    style B fill:#95e1d3
```

**Version Management:**
- Each update creates new version
- Current version tracked in metadata
- Old versions retained up to `max_versions`
- Can read any non-deleted version
- Oldest versions pruned automatically

**Example:**
```bash
# Get current version (5)
spike secret get secrets/db/creds

# Get specific version
spike secret get secrets/db/creds --version 3

# List versions
spike secret versions secrets/db/creds
# Output:
#   Version 5 (current) - 2024-01-20
#   Version 4 - 2024-01-15
#   Version 3 - 2024-01-10
#   Version 2 - 2024-01-05
#   Version 1 - 2024-01-01
```

---

## 5. Secret Soft Delete

```mermaid
sequenceDiagram
    participant User as Operator
    participant Pilot as SPIKE Pilot
    participant Nexus as SPIKE Nexus
    participant State as State Manager
    participant DB as SQLite Database

    Note over User: Soft delete secret

    User->>Pilot: spike secret delete secrets/db/creds

    Pilot->>Nexus: POST /v1/secret/delete<br/>SecretDeleteRequest<br/>{path, version (optional)}

    Nexus->>Nexus: Check authorization (delete permission)

    Nexus->>State: DeleteSecret(path, version)

    alt Version specified
        State->>DB: UPDATE secrets<br/>SET deleted_time = now<br/>WHERE path = ? AND version = ?
        Note right of DB: Soft delete specific version
    else Version not specified
        State->>DB: SELECT current_version<br/>FROM secret_metadata<br/>WHERE path = ?
        DB-->>State: current_version

        State->>DB: UPDATE secrets<br/>SET deleted_time = now<br/>WHERE path = ?<br/>AND version = current_version
        Note right of DB: Soft delete current version only
    end

    State-->>Nexus: Secret deleted

    Nexus-->>Pilot: 200 OK

    Pilot-->>User: Secret deleted

    Note over DB: Secret still in database<br/>but has deleted_time set.<br/>Queries exclude deleted secrets<br/>unless specifically requested.
```

**Key Files:**
- `app/spike/internal/cmd/secret/delete.go` - CLI command
- `app/nexus/internal/route/secret/delete.go` - HTTP handler

**Soft Delete Benefits:**
- Can be undeleted
- Audit trail preserved
- No data loss
- Fast operation (no encryption/decryption)

---

## 6. Secret Undelete

```mermaid
sequenceDiagram
    participant User as Operator
    participant Pilot as SPIKE Pilot
    participant Nexus as SPIKE Nexus
    participant State as State Manager
    participant DB as SQLite Database

    Note over User: Restore deleted secret

    User->>Pilot: spike secret undelete secrets/db/creds

    Pilot->>Nexus: POST /v1/secret/undelete<br/>SecretUndeleteRequest<br/>{path, version (optional)}

    Nexus->>Nexus: Check authorization (write permission)

    Nexus->>State: UndeleteSecret(path, version)

    alt Version specified
        State->>DB: UPDATE secrets<br/>SET deleted_time = NULL<br/>WHERE path = ? AND version = ?
    else Version not specified
        State->>DB: UPDATE secrets<br/>SET deleted_time = NULL<br/>WHERE path = ?<br/>AND version = current_version
    end

    State-->>Nexus: Secret undeleted

    Nexus-->>Pilot: 200 OK

    Pilot-->>User: Secret restored

    Note over DB: deleted_time cleared.<br/>Secret visible in queries again.
```

**Note:** Undelete only works for soft-deleted secrets (marked with
`deleted_time`). Once a version is physically removed from the database
through automatic pruning (when `MaxVersions` is exceeded and new versions
are added), it cannot be recovered. There is no explicit "hard delete" API
endpoint—physical deletion only occurs automatically during version pruning.

---

## 7. Secret List

```mermaid
sequenceDiagram
    participant User as Operator
    participant Pilot as SPIKE Pilot
    participant Nexus as SPIKE Nexus
    participant Policy as Policy Engine
    participant DB as SQLite Database

    Note over User: List all secrets

    User->>Pilot: spike secret list

    Pilot->>Nexus: POST /v1/secret/list<br/>SecretListRequest

    Nexus->>Nexus: Extract peer SPIFFE ID

    Nexus->>DB: SELECT DISTINCT path<br/>FROM secret_metadata

    DB-->>Nexus: All secret paths

    loop For each path
        Nexus->>Policy: CheckPermission(spiffeID, path, read)

        alt Has read permission
            Policy-->>Nexus: Authorized
            Note over Nexus: Include in results
        else No read permission
            Policy-->>Nexus: Unauthorized
            Note over Nexus: Exclude from results
        end
    end

    Nexus->>Nexus: Filter paths by policy

    Nexus-->>Pilot: 200 OK<br/>SecretListResponse<br/>{paths: [authorized paths]}

    Pilot-->>User: Display secret paths

    Note over User: Operator sees only secrets<br/>they have permission to read
```

**Key Files:**
- `app/spike/internal/cmd/secret/list.go` - CLI command
- `app/nexus/internal/route/secret/list.go` - HTTP handler

**Security Note:** List returns only paths, not secret data. Only paths the
user has permission to read are included.

---

## 8. Secret Version Pruning

```mermaid
sequenceDiagram
    participant Nexus as SPIKE Nexus
    participant State as State Manager
    participant Metadata as secret_metadata
    participant Secrets as secrets table

    Note over Nexus: During secret update

    Nexus->>State: UpsertSecret(path, values)

    State->>State: Insert new version

    State->>Metadata: SELECT max_versions,<br/>current_version,<br/>oldest_version<br/>WHERE path = ?

    Metadata-->>State: max_versions = 10<br/>current_version = 15<br/>oldest_version = 6

    State->>State: Calculate: versions = current - oldest
    Note right of State: versions = 15 - 6 = 10

    alt versions >= max_versions
        State->>State: Calculate prune threshold
        Note right of State: prune_below = current - max_versions<br/>prune_below = 15 - 10 = 5

        State->>Secrets: DELETE FROM secrets<br/>WHERE path = ?<br/>AND version < ?
        Note right of Secrets: Delete versions < 5<br/>(versions 1-4 removed)

        State->>Metadata: UPDATE secret_metadata<br/>SET oldest_version = ?<br/>WHERE path = ?
        Note right of Metadata: oldest_version = 5

        Note over State: Versions 5-15 retained.<br/>Versions 1-4 permanently deleted.
    else versions < max_versions
        Note over State: No pruning needed.<br/>All versions retained.
    end
```

**Pruning Rules:**
- Automatic during secret updates
- Keeps most recent `max_versions` versions
- Oldest versions physically deleted from database
- Cannot be recovered after pruning
- This is the only way versions are permanently removed (no manual hard
  delete API)

**Configuration:**
```bash
# Set max versions for new secrets (default: 10)
spike secret put path key=value --max-versions 20

# Update max versions for existing secret
spike secret metadata path --max-versions 5
```

---

## 9. Complete Secret Lifecycle State Machine

```mermaid
stateDiagram-v2
    [*] --> Created: spike secret put

    Created --> Updated: spike secret put<br/>(new version)
    Updated --> Updated: spike secret put<br/>(new version)

    Created --> Read: spike secret get
    Updated --> Read: spike secret get
    Read --> Created: Continue reading
    Read --> Updated: Continue reading

    Created --> SoftDeleted: spike secret delete
    Updated --> SoftDeleted: spike secret delete

    SoftDeleted --> Restored: spike secret undelete
    Restored --> Read: spike secret get

    Updated --> Pruned: Automatic pruning<br/>(old versions exceed MaxVersions)
    Pruned --> [*]: Permanent deletion<br/>(no recovery possible)

    note right of Created
        Version 1 created.
        Encrypted and stored.
    end note

    note right of Updated
        New version created.
        Old versions retained.
    end note

    note right of SoftDeleted
        deleted_time set.
        Can be undeleted.
    end note

    note right of Pruned
        Physical deletion (automatic).
        Oldest versions removed when
        MaxVersions exceeded.
        Cannot be recovered.
    end note
```

---

## 10. Secret Lifecycle Operations Summary

| Operation | Command | Effect | Reversible |
|-----------|---------|--------|------------|
| **Create** | `spike secret put path key=val` | Create version 1 | N/A |
| **Read** | `spike secret get path` | Get current version | N/A (read-only) |
| **Read Version** | `spike secret get path --version N` | Get specific version | N/A (read-only) |
| **Update** | `spike secret put path key=newval` | Create new version | Yes (old version retained) |
| **List** | `spike secret list` | List paths (filtered by policy) | N/A (read-only) |
| **Versions** | `spike secret versions path` | List all versions | N/A (read-only) |
| **Soft Delete** | `spike secret delete path` | Set deleted_time | Yes (undelete) |
| **Undelete** | `spike secret undelete path` | Clear deleted_time | N/A (restores) |
| **Prune** | Automatic (on update) | Physically removes old versions | No |

---

## 11. Key Files Reference

**CLI Commands:**
- `app/spike/internal/cmd/secret/put.go` - Create/update secret
- `app/spike/internal/cmd/secret/get.go` - Read secret
- `app/spike/internal/cmd/secret/delete.go` - Delete secret
- `app/spike/internal/cmd/secret/undelete.go` - Restore secret
- `app/spike/internal/cmd/secret/list.go` - List secrets
- `app/spike/internal/cmd/secret/versions.go` - List versions

**HTTP Handlers:**
- `app/nexus/internal/route/secret/put.go` - Secret put endpoint
- `app/nexus/internal/route/secret/get.go` - Secret get endpoint
- `app/nexus/internal/route/secret/delete.go` - Secret delete endpoint
- `app/nexus/internal/route/secret/undelete.go` - Secret undelete endpoint
- `app/nexus/internal/route/secret/list.go` - Secret list endpoint

**State Management:**
- `app/nexus/internal/state/base/secret.go` - Secret operations
- `app/nexus/internal/state/backend/sqlite/persist/secret.go` - Database persistence

---

## Summary

**Secret Lifecycle Stages:**
1. **Creation**: First version created
2. **Reading**: Access current or specific version
3. **Updating**: Create new version, retain old
4. **Versioning**: Track multiple versions per secret
5. **Soft Deletion**: Mark as deleted (reversible)
6. **Undelete**: Restore soft-deleted secret
7. **Pruning**: Automatic removal of old versions
8. **Hard Deletion**: Permanent removal (irreversible)

**Key Features:**
- **Versioning**: Full history of secret changes
- **Soft Delete**: Reversible deletion with undelete
- **Auto-Pruning**: Keep most recent versions, remove old
- **Policy-Based Access**: Authorization at every step
- **Encryption**: All versions encrypted with unique nonce
- **Audit Trail**: Timestamps for all operations
