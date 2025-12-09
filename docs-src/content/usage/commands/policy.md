+++
#    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
#  \\\\\ Copyright 2024-present SPIKE contributors.
# \\\\\\\ SPDX-License-Identifier: Apache-2.

title = "spike policy"
weight = 3
sort_by = "weight"
+++

# `spike policy`

The `spike policy` command is the main entry point for managing **access
policies** in SPIKE. It allows administrators to define, view, and manage rules
that control access to secrets and resources based on workload identity
(**SPIFFE ID**) and resource paths.

SPIKE provides two commands for managing policies:

1. **`spike policy create`**---Traditional command-line interface 
  (*backward compatibility*)
2. **`spike policy apply`**---Enhanced command with YAML file support 
  (*recommended for new workflows*)

While `spike policy create` checks for the existence of a policy, and
errors out if we are overriding an existing policy, `spike policy apply` uses
**upsert semantics**---it will create a new policy if one doesn't exist, or 
update an existing policy if one with the same name already exists. This makes 
the `spike policy apply` command safe to use in automation and GitOps workflows.

## Quick Start

```bash
# Using YAML file (recommended)
spike policy apply --file policy.yaml
```

## YAML File Format

### Basic Structure
```yaml
# Policy name - must be unique within the system
name: "web-service-policy"

# SPIFFE ID RegEx pattern for workload matching
spiffeidPattern: "^spiffe://example\\.org/web-service/$"

# Path RegEx pattern for access control
pathPattern: "^secrets/web-service/db-[0-9]*$"

# List of permissions to grant
permissions:
  - read
  - write
```

### Realistic SPIFFE ID Pattern and Path Pattern Examples

```yaml
# Database secrets
name: "database-policy"
spiffeidPattern: "^spiffe://example\\.org/database$"
pathPattern: "^secrets/database/production$"
permissions: [read]

# Web service configuration
name: "web-service-policy"
spiffeid: "^spiffe://example\\.org/web-service$"
path: "^secrets/web-service/config$"
permissions: [read, write]

# Cache credentials
name: "cache-policy"
spiffeid: "^spiffe://example\\.org/cache/$"
path: "^secrets/cache/redis/session$"
permissions: [read]

# Application environment variables
name: "app-env-policy"
spiffeid: "^spiffe://example\\.org/app$"
path: "^secrets/app/env/production$"
permissions: [read, list]
```

### All Available Permissions
```yaml
name: "admin-policy"
spiffeid: "^spiffe://example\\.org/admin$"
path: "secrets"
permissions:
  - read    # Permission to read secrets
  - write   # Permission to create, update, or delete secrets
  - list    # Permission to list resources
  - super   # Administrative permissions
```

### Alternative YAML Formats

#### Flow Sequence for Permissions
```yaml
name: "database-policy"
spiffeid: "^spiffe://example\\.org/database$"
path: "^secrets/database/production$"
permissions: [read, write, list]
```

#### Quoted Values
```yaml
name: "cache-policy"
spiffeidPattern: "^spiffe://example\\.org/cache$"
pathPattern: "^secrets/cache/redis$"
permissions:
  - "read"
  - "write"
```

## Creating Policies Using Command-Line Flags

Instead of using a `yaml` file, you can provide command-line arguments
to programmatically create your policies too:

```bash
# Create your first policy
spike policy create --name=my-service \
  --path-pattern="^secrets/app$" \
  --spiffeid-pattern="^spiffe://example\.org/service$" \
  --permissions=read

# Verify your policy was created
spike policy list
```

## What are SPIKE Policies?

Policies in **SPIKE** provide a secure and flexible way to control access to 
secrets and resources. Each policy defines:

* **Who** can access resources (via **SPIFFE ID** patterns)
* **What** resources can be accessed (via **path** patterns)
* **How** resources can be accessed (via **permissions**)

Policies are the cornerstone of **SPIKE**'s security model, allowing for 
fine-grained access control based on workload identity. Using 
[**SPIFFE ID**][spiffe-concepts]s as the foundation, **SPIKE** ensures that 
only authorized workloads can access sensitive information.

[spiffe-concepts]: https://spiffe.io/docs/latest/spiffe-about/spiffe-concepts/ "SPIFFE Concepts"

### How Policies Work

When a workload attempts to access a resource in SPIKE:

1. The workload presents its **SPIFFE ID** through a 
   SPIFFE Verifiable Identity Document (**SVID**)
2. **SPIKE** validates the **SVID** to verify the workload's identity
3. **SPIKE** checks if any policy matches both:
   * The workload's **SPIFFE ID** against the policy's **SPIFFE ID pattern**
   * The requested **resource path** against the policy's **path pattern**
4. If a match is found, SPIKE checks if the requested operation is allowed by 
   the policy's **permissions**
5. Access is granted only if **ALL** conditions are met

### Why Use Policies?

* **Zero Trust Security**: Access is based on workload identity, not network 
  location
* **Least Privilege**: Grant only the permissions needed for each workload
* **Auditability**: All access is tied to specific policies and identities
* **Flexibility**: Patterns support regular expression matching, which allows
  a more fine-grained control over which resources the policy applies to.
* **Scalability**: Policies work consistently across any deployment size

## Features

* **Create policies** with specific permissions and access patterns
* **Apply policies** using upsert semantics (create new or update existing)
* **List all policies** in human-readable or JSON format
* **Get policy details** by ID or name
* **Delete policies** with confirmation protection
* **Enhanced validation** for permissions and parameters

## Commands

### `spike policy list`

```bash
spike policy list [--format=human|json] [--path-pattern=<pattern> | --spiffeid-pattern=<pattern>]
```

Lists all policies in the system. Can be filtered by a resource path pattern or 
a SPIFFE ID pattern.

When using filters, you must provide **the exact regular expression pattern** as
defined in the policies you want to match. For example, if a policy is defined
with pattern `^secrets/database/production$`, you must use exactly that pattern
to find it---no partial matches or simpler patterns will work.

**Note:** `--path-pattern` and `--spiffeid-pattern` flags cannot be used 
together.

### `spike policy create`

```bash
spike policy create --name=<name> \
  --path-pattern=<path-pattern> \
  --spiffeid-pattern=<spiffe-id-pattern> \
  --permissions=<permissions>
```
Creates a new policy with the specified parameters.

### `spike policy apply`

```bash
spike policy apply --file=<policy-file.yaml>
```

Creates a new policy with file-based input using YAML configuration.

#### YAML Configuration Format

When using the `--file` flag, the YAML file should follow this structure:

```yaml
name: policy-name
spiffeidPattern: ^spiffe://example\.org/service$
pathPattern: ^secrets/database/production$
permissions:
  - read
  - write
```

### Example Files

SPIKE repository has the following example policies for your convenience:

* [`./examples/policies/sample-policy.yaml`][policy-example]---Basic policy example
* [`./examples/policies/test-policies/basic-policy.yaml`][basic-policy]---Minimal 
  policy
* [`./examples/policies/test-policies/admin-policy.yaml`][admin-policy]---Full 
  permissions policy
* [`./examples/policies/test-policies/invalid-permissions.yaml`][invalid-perms]---Example 
  with invalid permissions (for testing)

[policy-example]: https://github.com/spiffe/spike/blob/main/examples/policies/sample-policy.yaml
[basic-policy]: https://github.com/spiffe/spike/blob/main/examples/policies/test-policies/basic-policy.yaml
[admin-policy]: https://github.com/spiffe/spike/blob/main/examples/policies/test-policies/admin-policy.yaml
[invalid-perms]: https://github.com/spiffe/spike/blob/main/examples/policies/test-policies/invalid-permissions.yaml

#### Permission Types

| Permission | Description                                            |
|------------|--------------------------------------------------------|
| **read**   | Allows reading secrets and resources                   |
| **write**  | Allows creating, updating, and deleting secrets        |
| **list**   | Allows listing resources and directories               |
| **super**  | Full administrative permissions (**use with caution**) |

#### Validation

All policy configurations are validated to ensure:

1. **Required fields**: `name`, `spiffeid`, `path`, and `permissions` must be 
   present
2. **Valid permissions**: Only `read`, `write`, `list`, and `super` are allowed
3. **Valid YAML syntax**: Proper YAML formatting is required (for YAML files)
4. **Non-empty values**: All fields must have non-empty values

#### GitOps Integration

YAML files can be easily integrated into GitOps workflows:

1. **Store policy YAML files in a Git repository**
   ```txt
   policies/
   ├── web-service-policy.yaml
   ├── database-policy.yaml
   └── admin-policy.yaml
   ```

2. **Use CI/CD pipelines to validate policies before deployment**
   ```bash
   # Validation step in CI
   for policy in policies/*.yaml; do
     spike policy apply --file "$policy"
     # - ensure that the policy is created
     # - delete the policy
     # - ensure that the policy is gone
   done
   ```

3. **Apply policies using `spike policy apply --file` in deployment scripts**
   ```bash
   # Deployment script
   for policy in policies/*.yaml; do
     spike policy apply --file "$policy"
   done
   ```
4. **Version control changes to policies alongside application code**

5. **Use upsert semantics to safely apply policy changes without worrying 
   about conflicts**

### `spike policy get`

```bash
spike policy get <id> [--format=human|json]
spike policy get --name=<name> [--format=human|json]
```

Gets details of a specific policy by ID or name. Use `--format=json` 
for machine-readable output.

### `spike policy delete`

```bash
spike policy delete <id>
spike policy delete --name=<name>
```

Deletes a policy by ID or name. Requires confirmation.

## Usage Examples

```bash
# Create a policy for a web service with read and write access
spike policy create \
  --name=web-service \
  --path-pattern="^secrets/web$" \
  --spiffeid-pattern="^spiffe://example\.org/web$" \
  --permissions=read,write

# Create a policy with multiple permissions
spike policy create \
  --name=admin-service \
  --path-pattern="secrets/" \
  --spiffeid-pattern="^spiffe://example\.org/admin$" \
  --permissions=read,write,list

# Apply a policy using a YAML file
spike policy apply --file=policy.yaml

# List all policies in JSON format (useful for automation)
spike policy list --format=json

# Get details of a specific policy by name
spike policy get --name=web-service

# Get policy details in JSON format
spike policy get --name=web-service --format=json

# Delete a policy and confirm deletion
spike policy delete --name=web-service
```

## Pattern Syntax

**SPIKE** policies support **regular expression** pattern matching for both 
SPIFFE IDs and resource paths:

- The pattern is compiled as a "*regular expression*".

This would mean, for an exact match, you would need to include `^` and `$` in
your patterns as well.

For example:

* `secrets/db` matches `global/secrets/db` and `secrets/db/local`
* Whereas, `^secrets/db$` only matches `secrets/db` and nothing else 
  (*`global/secrets/db` and `secrets/db/local` will not match*)

Thus, for precise control, you are encouraged to include `^` and `$` at the 
beginning and end of your patterns respectively for an exact match.

## How Regular Expressions are Used For Policy Matching

More specifically, **SPIKE** compiles **SPIFFE ID patterns** and 
**path patterns** defined in the policies into **regular
expressions**. 

Here is a simplified version of how this regular expression compilation
happens behind-the-scenes:

```go
pathRegex, err := regexp.Compile(policy.PathPattern)
// ... error handling omitted for brevity.
policy.PathRegex = pathRegex
// `pathRegEx` is used for policy validation.
```

As seen from the example above, both the path pattern and the SPIFFE ID
pattern that are provided during policy creation which are used "**AS IS**" to 
create regular expression matchers. These patterns are compiled
into Go's built-in regex engine, ensuring that the matching process strictly
adheres to the patterns defined in the policy, allowing for precise and flexible
access control.

## Simplicity Is the Key

Because of the regular expression usage in SPIKE policies, a `policy create` 
operation can define more flexible matching patterns. However, keeping patterns 
simple is both more secure and easier to manage and reason about. Creating a 
pattern that is too broad or that uses overly complex regular expressions may 
lead to unintended consequences and security risks. **Simplicity** is important 
to ensure patterns are clear, predictable, and effective.

When a workload attempts to access a resource, its **SPIFFE ID** and the 
requested resource **path** are matched against these compiled **regular
expressions**. This ensures that both identity and resource patterns follow the
specified rules and allow for flexibility with wildcards or exact matches.

### Path Pattern Examples

```txt
^secrets/               # All resources in the secrets directory
^secrets/database/      # Only resources in the database subdirectory  
^secrets/database/creds # Only the specific creds resource

# You can provide regular expressions for a more fine-tuned
# pattern match:
^secrets/db-[123]$ # Matches secrets/db-2, but not secrets/db-4.
```

### SPIFFE ID Pattern Examples
```
^spiffe://example\.org/             # Workloads in the example.org trust domain
^spiffe://example\.org/web/         # Only web workloads
^spiffe://example\.org/web/server$  # Only the specific web server workload
```

## Best Practices

* Follow the principle of least privilege when assigning permissions
* Use descriptive policy names that reflect their purpose
* Create separate policies for different workload types
* Use specific path patterns rather than overly broad ones
* Regularly audit and review your policies
* Never assign `super` permissions unless absolutely necessary

## Technical Details

### Permission Hierarchy

The `super` permission acts as a wildcard that grants all other permissions:

| Permission  | Description                            |
|-------------|----------------------------------------|
| `super`     | All permissions (wildcard)             |
| `write`     | Create and update secrets              |
| `read`      | Read secrets                           |
| `list`      | List secret paths                      |
| `execute`   | Cipher operations (encrypt/decrypt)    |

### Authorization for Policy Management

Policy management operations (create, update, delete) are authorized as follows:

1. **SPIKE Pilot** (`spiffe://<trustRoot>/spike/pilot/*`) has full access to
   all operations, including policy management
2. **Other workloads** need a policy granting `write` permission on the
   system path `spike/system/acl`

### Encryption at Rest

Policy details are encrypted in the database using **AES-256-GCM**:

**Encrypted fields:**
* SPIFFE ID Pattern (regex string)
* Path Pattern (regex string)
* Permissions (JSON array)

**Not encrypted:**
* Policy name (used for lookups)
* Policy ID
* Timestamps

A single nonce is generated per policy and used for all encrypted fields
to ensure atomicity.

### Policy Evaluation

When a secret is accessed, **SPIKE Nexus** evaluates policies by:

1. Checking if the requestor is **SPIKE Pilot** (grants immediate access)
2. Loading all policies from the backing store
3. For each policy, checking if the SPIFFE ID pattern matches the requestor
4. If matched, checking if the path pattern matches the requested resource
5. If matched, checking if the policy grants the required permission
6. Access is granted on **first match**; there are no "deny" policies

Policies are loaded fresh from the database on each request to ensure
changes take effect immediately.

### Regex Safety

SPIKE uses Go's `regexp` package which provides **linear-time** matching
guarantees. This prevents ReDoS (Regular Expression Denial of Service)
attacks.

## Common Errors

**Pattern validation failed:**
```
Error: Invalid SPIFFE ID pattern: "spiffe://example.org/workload/*"
Use regex syntax: "spiffe://example\.org/workload/.*"
```

**Unauthorized:**
```
Error: Permission denied
Only SPIKE Pilot or workloads with write access to spike/system/acl
can manage policies
```

**Path starts with a slash:**
```
Error: Invalid path pattern: "/secrets/app/.*"
Paths are namespaces, remove leading slash: "secrets/app/.*"
```

**Empty policy name:**
```
Error: Policy name cannot be empty
```

----

## `spike` Command Index

{{ toc_commands() }}

<p>&nbsp;</p>

----

{{ toc_usage() }}

----

{{ toc_top() }}
