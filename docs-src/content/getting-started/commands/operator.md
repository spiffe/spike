+++
# //    \\ SPIKE: Secure your secrets with SPIFFE.
# //  \\\\\ Copyright 2024-present SPIKE contributors.
# // \\\\\\\ SPDX-License-Identifier: Apache-2.

title = "spike operator"
weight = 4
sort_by = "weight"
+++

# `spike operator`

The `spike operator` command provides administrative functionality for 
**disaster recovery** and **system restoration** in **SPIKE**. It allows 
privileged operators with specific SPIFFE roles to perform critical operations 
for maintaining **SPIKE Nexus**' resilience.

## Quick Start

```bash
# For disaster recovery preparation (while system is healthy)
spike operator recover

# For system restoration (after system failure)
spike operator restore
```

## What is SPIKE Operator Mode?

The **Operator** mode in **SPIKE** provides privileged functionality for system 
maintenance and recovery. These commands:

* Are restricted to users with specific SPIFFE roles (`recover` or `restore`)
* Handle sensitive security operations for disaster recovery
* Manage the cryptographic shards needed for system restoration
* Provide secure mechanisms for recovering from catastrophic system failures

Operator commands are the safety net for **SPIKE** installations, ensuring that 
even in worst-case scenarios, the system can be recovered without compromising 
security.

## Commands

### `spike operator recover`

```bash
spike operator recover
```

The `recover` command allows privileged operators with the `recover` role to 
extract recovery shards from a healthy SPIKE Nexus system. These shards are 
essential for system restoration in case of catastrophic failure.

#### Requirements:

* Caller must have the `recover` SPIFFE role
* **SPIKE Nexus** must be running and healthy
* A recovery directory must be configured and accessible

#### Process:

1. Authenticates the caller's SPIFFE ID for the `recover` role
2. Retrieves recovery shards from the SPIKE API
3. Cleans the recovery directory of any previous recovery files
4. Saves the retrieved shards as text files in the recovery directory
5. Provides instructions for securing the recovery shards

#### Security Considerations:

* Recovery shards are security-critical and must be protected
* After recovery, shards should be encrypted and securely stored
* The recovery directory should be cleaned after shards are secured
* Loss of recovery shards may prevent system restoration

#### Example:

```bash
# Generate recovery shards for disaster preparedness
spike operator recover
```

After executing the command, you will see:

```
SPIKE Recovery shards saved to the recovery directory:
/path/to/recovery/directory

Please make sure that:
  1. You encrypt these shards and keep them safe.
  2. Securely erase the shards from the recovery directory after you encrypt them
     and save them to a safe location.

If you lose these shards, you will not be able to recover SPIKE Nexus in the
unlikely event of a total system crash.
```

### `spike operator restore`

```bash
spike operator restore
```

The `restore` command allows privileged operators with the `restore` role to 
restore **SPIKE Nexus** after a system failure. It requires the recovery shards 
previously generated with the `recover` command.

#### Requirements:

* Caller must have the `restore` SPIFFE role
* SPIKE Nexus must be in a state that requires restoration
* Recovery shards must be available

#### Process:

1. Authenticates the caller's SPIFFE ID for the `restore` role
2. Prompts for a recovery shard (input is hidden for security)
3. Validates and processes the provided shard
4. Reports the current restoration status
5. May require multiple executions with different shards to complete restoration

#### Security Considerations:

* Recovery shards are security-critical and handled with care
* Input is hidden during shard entry to prevent exposure
* Recovery shards are cleared from memory after use
* The restoration process is designed to require multiple shards for security

#### Example:

```bash
# Begin restoration process
spike operator restore
```

During execution, you will be prompted:

```
(your input will be hidden as you paste/type it)
Enter recovery shard: 
```

After providing a valid shard, you will see one of two responses:

If restoration is complete:

```
SPIKE is now restored and ready to use.
Please run `./hack/spire-server-entry-su-register.sh` with necessary privileges
to start using SPIKE as a superuser.
```

If more shards are needed:

```
Shards collected: 1
Shards remaining: 2
Please run `spike operator restore` again to provide the remaining shards.
```

## Recovery Shard Format

Recovery shards follow a specific format:

```
spike:INDEX:HEXDATA
```

Where:

* `INDEX` is the numeric index of the shard
* `HEXDATA` is the 64-character hexadecimal representation of a 32-byte secret

The system enforces strict validation of this format to ensure security and 
proper restoration.

## Best Practices

* **Regular Recovery Preparation**: Periodically run `recover` on healthy 
  systems to ensure up-to-date shards
* **Secure Shard Storage**: Encrypt recovery shards and store them in secure, 
  separate locations. **DO NOT STORE SHARDS ON DISK UNENCRYPTED**, use a secure
  storage tool, like a password manager.
* **Access Control**: Strictly limit access to the `recover` and `restore` roles
* **Documentation**: Maintain secure documentation of recovery procedures
* **Testing**: Regularly test the recovery process in non-production environments
* **Multiple Administrators**: Distribute recovery shards among multiple trusted 
  administrators

## Security Considerations

* Recovery shards provide full system access and must be protected accordingly
* The system uses cryptographic techniques to secure recovery operations
* Memory containing shards is explicitly cleared after use
* Both commands implement role-based access control through SPIFFE IDs
* Recovery files are created with restrictive permissions (0600)

## Role Assignment

To assign the required roles for operator commands:

1. For recovery role:
   ```bash
   ./hack/spire-server-entry-recover-register.sh
   ```

2. For restore role:
   ```bash
   ./hack/spire-server-entry-restore-register.sh
   ```

These scripts must be run with appropriate privileges.

----

## `spike` Command Index

{{ toc_commands() }}

----

{{ toc_getting_started() }}

----

{{ toc_top() }}
