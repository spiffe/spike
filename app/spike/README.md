![SPIKE](../../assets/spike-banner-lg.png)

## SPIKE Pilot

**SPIKE Pilot** is the command-line interface for the **SPIKE** system.

It is a binary named `spike`.

It is helpful to define an alias to `spike` for ease of use:

```bash
# ~/.bashrc

# Define an alias to where your `spike` binary is:
alias spike=$HOME/WORKSPACE/spike-git-repo/spike
```

## Getting Help

Simply typing `spike` will show a summary of available commands.

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
  delete      Soft-delete a secret (can be recovered)
  undelete    Restore a soft-deleted secret
  metadata    Retrieve secret metadata
```

**Examples:**

```bash
spike secret put secrets/db/password username=admin password=secret
spike secret get secrets/db/password
spike secret get secrets/db/password -v 2          # Get specific version
spike secret get secrets/db/password username      # Get specific key
spike secret list
spike secret delete secrets/db/password
spike secret delete secrets/db/password -v 1,2,3   # Delete specific versions
spike secret undelete secrets/db/password
spike secret metadata get secrets/db/password
```

### Policy Management

```text
spike policy <subcommand>

Subcommands:
  create      Create a new policy
  apply       Apply a policy from a YAML file
  list        List all policies
  get         Get details of a specific policy by ID or name
  delete      Delete a policy by ID or name
```

**Examples:**

```bash
spike policy list
spike policy get abc123
spike policy get --name=my-policy
spike policy create --name=new-policy \
    --path-pattern="^secrets/.*$" \
    --spiffeid-pattern="^spiffe://example\.org/.*$" \
    --permissions=read,write
spike policy apply -f policy.yaml
spike policy delete abc123
spike policy delete --name=my-policy
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
spike cipher encrypt --in secret.txt --out secret.enc
spike cipher decrypt --in secret.enc --out secret.txt
echo "sensitive data" | spike cipher encrypt | spike cipher decrypt
```

### Operator Functions

```text
spike operator <subcommand>

Subcommands:
  recover     Recover Shamir shards from SPIKE Keeper
  restore     Restore SPIKE Nexus using recovery shards
```

These commands are used for disaster recovery scenarios when SPIKE Nexus
needs to be restored from Shamir secret shards.

## Notes

- Secret paths are namespace identifiers (e.g., `secrets/db/password`),
  not filesystem paths. They should **not** start with a forward slash.
- Policy patterns use **regex**, not globs (e.g., `^secrets/.*$` not
  `secrets/*`).
- The CLI is a constant work in progress, so what you see above might be
  slightly different from the version that you are using.

For additional help, [check the official documentation](https://spike.ist/).