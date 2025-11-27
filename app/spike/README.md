![SPIKE](../../assets/spike-banner-lg.png)

## SPIKE Pilot

**SPIKE Pilot** is the command-line interface for the **SPIKE** system.

The binary is named `spike`. You can define an alias for convenience:

```bash
# ~/.bashrc
alias spike=$HOME/WORKSPACE/spike/spike
```

## Getting Help

Running `spike` without arguments shows available commands:

```text
Usage: spike [command] [flags]

Commands:
  secret      Manage secrets
  policy      Manage policies
  cipher      Encrypt and decrypt data using SPIKE Nexus
  operator    Manage admin operations
```

## Command Groups

### Secret Management

```text
spike secret <subcommand>

Subcommands:
  put         Create or update a secret
  get         Retrieve a secret value
  list        List all secret paths
  delete      Soft-delete secret versions (can be recovered)
  undelete    Restore soft-deleted versions
  metadata    Retrieve secret metadata (versions, timestamps)
```

**Examples:**

```bash
spike secret put secrets/db/creds username=admin password=secret
spike secret get secrets/db/creds
spike secret get secrets/db/creds --version 2      # Get specific version
spike secret get secrets/db/creds username         # Get specific key
spike secret list
spike secret list secrets/db                       # Filter by prefix
spike secret delete secrets/db/creds
spike secret delete secrets/db/creds --versions 1,2,3
spike secret undelete secrets/db/creds --versions 1,2
spike secret metadata get secrets/db/creds
```

### Policy Management

```text
spike policy <subcommand>

Subcommands:
  create      Create a new policy
  apply       Create or update a policy (upsert)
  list        List all policies
  get         Get details of a specific policy
  delete      Delete a policy
```

**Examples:**

```bash
spike policy list
spike policy get abc123
spike policy get --name my-policy
spike policy create --name new-policy \
    --path-pattern "^secrets/.*$" \
    --spiffeid-pattern "^spiffe://example\.org/.*$" \
    --permissions read,write
spike policy apply --file policy.yaml
spike policy delete abc123
spike policy delete --name my-policy
```

### Cipher Operations

```text
spike cipher <subcommand>

Subcommands:
  encrypt     Encrypt data via SPIKE Nexus
  decrypt     Decrypt data via SPIKE Nexus
```

**Examples:**

```bash
# Stream mode (file or stdin/stdout)
spike cipher encrypt --file secret.txt --out secret.enc
spike cipher decrypt --file secret.enc --out secret.txt
echo "sensitive data" | spike cipher encrypt | spike cipher decrypt

# JSON mode (base64 input/output)
spike cipher encrypt --plaintext $(echo -n "secret" | base64)
spike cipher decrypt --ciphertext <base64> --nonce <base64> --version <n>
```

### Operator Functions

```text
spike operator <subcommand>

Subcommands:
  recover     Retrieve recovery shards (while SPIKE Nexus is healthy)
  restore     Submit recovery shards (to restore a failed SPIKE Nexus)
```

These commands are for disaster recovery when SPIKE Nexus cannot auto-recover
via SPIKE Keeper. See https://spike.ist/operations/recovery/ for details.

## Notes

- Secret paths are namespace identifiers (e.g., `secrets/db/password`),
  not filesystem paths. They should **not** start with a forward slash.
- Policy patterns use **regex**, not globs (e.g., `^secrets/.*$` not
  `secrets/*`).

For additional help, see the [official documentation](https://spike.ist/).