+++
# //    \\ SPIKE: Secure your secrets with SPIFFE.
# //  \\\\\ Copyright 2024-present SPIKE contributors.
# // \\\\\\\ SPDX-License-Identifier: Apache-2.

title = "SPIKE Backup and Restore"
weight = 4
sort_by = "weight"
+++

# SPIKE Backup and Restore Guide

**SPIKE**, the Secure Production Identity Framework for Key Encryption, provides 
robust secrets management with strong identity controls. But even the most 
reliable systems need comprehensive backup and recovery plans. This guide
explains how to properly back up, secure, and restore SPIKE deployments—ensuring 
your critical secrets infrastructure remains resilient against catastrophic 
failures.

## Architecture Foundations for Effective Backup Planning

SPIKE consists of three critical components, each requiring specific backup 
considerations:

1. **SPIKE Nexus**: The central component handling secret encryption/decryption 
   and root key management. It stores encrypted secrets in a SQLite database 
   and communicates via an mTLS API.

2. **SPIKE Keeper**: A redundancy mechanism holding Shamir Secret Sharing shards 
   of the root key in memory. Multiple Keeper instances provide resiliency, 
   requiring a configurable threshold of shards to reconstruct the root key.

3. **SPIKE Pilot**: The CLI interface that translates administrative commands 
   into mTLS API calls to **SPIKE Nexus**.

The foundation of SPIKE's security model lies in its root key management:

* The **root key** encrypts all data in the backing store and **never** exists 
  on disk in plaintext
* The system splits the root key into multiple Shamir shards based on a 
  configurable threshold (*e.g., any 2 of 3 or 3 of 5 shards needed to 
  reconstruct*)
* **SPIKE Keeper**s hold these shards in memory for automatic recovery if 
  **SPIKE Nexus** crashes
* For disaster recovery scenarios, administrators can create additional recovery 
  shards

## Backup procedures

### SQLite database backup

The SPIKE Nexus uses a SQLite database to store all encrypted secrets and 
metadata. This database is typically located in `~/.spike` on the Nexus server.

```bash
# 1. First, ensure consistent state by using SQLite's online backup API
sqlite3 ~/.spike/database.sqlite \
  ".backup '/backup/spike_nexus_db_$(date +%Y%m%d_%H%M%S).sqlite'"

# 2. For WAL mode databases, checkpoint first to ensure consistency
sqlite3 ~/.spike/database.sqlite "PRAGMA wal_checkpoint(FULL);"
sqlite3 ~/.spike/database.sqlite \
  ".backup '/backup/spike_nexus_db_$(date +%Y%m%d_%H%M%S).sqlite'"

# 3. Verify backup integrity
sqlite3 /backup/spike_nexus_db_*.sqlite "PRAGMA integrity_check;"
```

**Important considerations:**

* The SQLite database backup contains encrypted data that can only be decrypted 
  with the root key
* Use database-level locking through SQLite's `.backup` command rather than 
  direct file copying

### Root key and cryptographic material backup

The root key is **SPIKE**'s most critical component. While it exists only in 
memory during normal operation, you must back it up for disaster recovery using 
Shamir's Secret Sharing:

```bash
# Create recovery shards of the root key
# IMPORTANT: Run this BEFORE any disaster occurs
spike recover

# This will generate multiple shard files under `~/.spike/recovery` folder. 
```

**Secure handling of recovery shards:**

1. Encrypt each shard immediately after creation (*e.g., using GPG with 
   hardware keys*)
2. Store encrypted shards in separate secure locations
3. Consider using HSMs or smart cards for shard storage
4. Implement strict access controls for shard access
5. Document the threshold configuration (*e.g., "2 of 3 shards required"*)

### Configuration and other components backup

Beyond the database and root key, back up these critical components:

1. **SPIRE Server** and **SPIRE Agent** configuration.

2. SPIFFE registration entries:
   ```bash
   # Back up SPIFFE registration entries
   spire-server entry show > /backup/spire_entries_$(date +%Y%m%d).txt
   ```

## Restore procedures

### Prerequisites for Restoration

Before beginning any restore operation, ensure:

1. You have all necessary components:
   - SQLite database backup
   - Access to the required number of recovery shards (*meeting your threshold*)
   - SPIFFE/SPIRE configuration backups

2. You have the appropriate SPIFFE identity for restoration:
   - Required SPIFFE ID: `spiffe://$trustRoot/spike/pilot/role/restore`

3. All **SPIKE** services are properly installed on the target system

### Root key restoration

If both **SPIKE Nexus** and all **SPIKE Keeper**s are unavailable 
(*catastrophic failure*), follow this procedure:

```bash
# 1. Configure SPIKE Pilot for restore operations 
# (adjust the script path for your environment)
./hack/spire-server-entry-restore-register.sh

# 2. Run the restore command
spike restore

# 3. When prompted, provide recovery shards one at a time
# You'll need to provide enough shards to meet your threshold (e.g., 2 of 3)

# 4. After successful restoration, revert SPIKE Pilot to normal operation
./hack/spire-server-entry-su-register.sh
```

SPIKE Nexus will:
* Automatically reconstruct the root key from the provided shards
* Redistribute shards to available SPIKE Keeper instances
* Resume normal operation with the restored key

### SQLite database restoration

To restore the SQLite database:

1. Stop **SPIKE Nexus**.

2. Replace the current database with the backup.
   ```bash
   cp /backup/spike_nexus_db_TIMESTAMP.sqlite \
     ~/.spike/database.sqlite

3. Set appropriate permissions
   ```bash
   chown spike:spike ~/.spike/database.sqlite
   chmod 600 ~/.spike/database.sqlite
   ```

4. Start **SPIKE Nexus**

**Note**: After restoring the database, if **SPIKE Nexus** cannot automatically 
recover the root key from **SPIKE Keeper**s, you'll need to perform the root 
key restoration procedure above.

### Verification procedures

After completing a restore operation, verify system integrity:

```bash
# Verify database integrity
sqlite3 ~/.spike/database.sqlite "PRAGMA integrity_check;"

# Test secret access to verify encryption/decryption is working
spike get /path/to/test/secret
```

## Backup best practices

### Backup frequency and scheduling

| Component       | Recommended Frequency                               | Reasoning                                |
|-----------------|-----------------------------------------------------|------------------------------------------|
| SQLite Database | Daily                                               | Captures secret changes promptly         |
| Root Key Shards | After initial setup and after any root key rotation | Critical security component              |
| Configuration   | After any configuration change                      | Ensures you can recreate the environment |
| SPIFFE Entries  | After any identity changes                          | Required for workload authentication     |

### Backup rotation and retention

Implement a comprehensive retention policy:

* **Short-term backups**: Keep daily backups for 14 days
* **Medium-term backups**: Keep weekly backups for 3 months
* **Long-term backups**: Keep monthly backups for 1 year

> **Test Your Backup Integrity**
> 
> A backup that does not work when you need most is not a backup.
> Make sure you validate the integrity and efficacy of your backups
> regularly.

### Secure Storage Recommendations

For root key recovery shards:

* **Multi-level security**: Encrypt shards before storage
* **Physical separation**: Store shards in different physical locations
* **Access controls**: Implement strict controls with separation of duties
* **Hardware security**: Consider HSMs or smart cards for shard storage
* **Environmental protection**: Use fire/water-resistant safes for physical media

For database backups:

- **Encryption**: Implement at-rest encryption for all backup files
- **Access limitations**: Restrict backup access to authorized personnel only
- **Immutability**: Consider WORM (Write Once Read Many) storage for critical backups
- **Offline copies**: Maintain air-gapped copies of critical backups

## Migration Limitations and Planning

While **SPIKE** currently lacks built-in migration tools, careful planning can 
facilitate future migrations.

### Current migration limitations

- No direct database migration between different SPIKE versions
- Manual coordination is required for root key transfers

## Conclusion

A comprehensive backup and restore strategy is essential for maintaining the 
resilience of your **SPIKE** deployment. By following the procedures in this 
guide, you can ensure that even in catastrophic failure scenarios, your secrets 
management infrastructure can be rapidly restored with minimal data loss.

Remember these key principles:

* **Regular backups**: Automated, validated, and securely stored
* **Root key protection**: The foundation of your security model
* **Tested procedures**: Verify your restore process works before you need it
* **Documentation**: Keep clear records of all configurations and procedures

By implementing these practices, your DevOps team will be prepared to handle 
any recovery scenario while maintaining the security guarantees that make 
**SPIKE** an effective secrets management solution.
