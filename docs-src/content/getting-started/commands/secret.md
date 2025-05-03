+++
# //    \\ SPIKE: Secure your secrets with SPIFFE.
# //  \\\\\ Copyright 2024-present SPIKE contributors.
# // \\\\\\\ SPDX-License-Identifier: Apache-2.

title = "spike secret"
weight = 2
sort_by = "weight"
+++

# `spike secret`

The `spike secret` command is the main entry point for managing **secrets** in 
**SPIKE**. It allows administrators to create, read, update, and delete secrets 
based on SPIFFE identities and corresponding access policies.

## Quick Start

```bash
# Store a new secret
spike secret put secrets/app/config api_key=abc123 environment=production

# Retrieve a secret
spike secret get secrets/app/config

# List all secrets
spike secret list

# Delete a secret version
spike secret delete secrets/app/config
```

## What are SPIKE Secrets?

Secrets in **SPIKE** are sensitive pieces of information that need to be 
securely stored, accessed, and managed. Each secret:

* Is stored at a specific **path**
* Contains one or more **key-value pairs**
* Has **version history** for auditing and recovery
* Is protected by **access policies** based on workload identity

Secrets are the core data objects managed by **SPIKE**, providing a secure way 
to distribute sensitive configuration data, credentials, and other confidential 
information to authorized workloads based on their SPIFFE identities.

### How Secrets Work

When a workload attempts to access a secret in SPIKE:

1. The workload presents its **SPIFFE ID** through a SPIFFE Verifiable Identity 
   Document (**SVID**)
2. **SPIKE** validates the **SVID** to verify the workload's identity
3. **SPIKE** checks if any policy allows the workload to access the requested 
   secret path
4. If authorized, the secret is securely delivered to the workload

This ensures that only authorized workloads can access specific secrets based on 
their verified identity, following zero-trust security principles.

## Path Syntax and Conventions

Secret paths in **SPIKE** have specific syntax requirements and recommended 
conventions to ensure consistency and avoid common pitfalls.

### Path Format Requirements

All secret paths must match the regex pattern:

```
^[a-zA-Z0-9._\-/()?+*|[\]{}\\]+$
```

This pattern allows alphanumeric characters, dots, underscores, hyphens, forward 
slashes, parentheses, question marks, plus signs, asterisks, pipes, square 
brackets, curly braces, and backslashes.

### Path Format Recommendations

While the validation requirements allow for flexibility, the following 
conventions are strongly recommended:

* **Avoid leading slashes**: Paths should not start with a forward slash (`/`)
* **Use forward slashes** to create hierarchical structures (like a file system)
* **Use descriptive, hierarchical naming** to organize secrets logically
* **Avoid double slashes** or other ambiguous path constructions
* **Avoid special characters** when possible, even if they are technically allowed

### Example Valid Paths

* ✅ `secrets/myapp/config` - Clear hierarchy, no leading slash  
* ✅ `secrets/db-creds/admin-user` - Well-structured with hyphens  
* ✅ `tenantA/projectX/env1/key` - Multi-level organization

### Example Invalid or Discouraged Paths

* ❌ `/secrets/myapp/config` - Avoid leading slashes  
* ❌ `secrets//double-slash` - Avoid double slashes  
* ❌ `secret\path` - Avoid backslashes (use forward slashes)  
* ❌ `secret path/with space` - Avoid spaces  
* ❌ `secret#invalid?path` - Avoid URL-reserved characters when possible

### Best Path Practices

* Use consistent prefixes like `secrets/` or `credentials/` as the first segment
* Organize paths by application, service, or environment
* Include version indicators in the path for managed rotation (e.g., `secrets/database/v1/credentials`)
* Use clear, descriptive names that indicate the purpose of the secret
* Keep paths reasonably short while maintaining clarity

### Path Examples

```
secrets/app/config                # Application configuration
secrets/database/production/creds # Production database credentials
secrets/certificates/tls          # TLS certificates
secrets/api/external/stripe/key   # External API credentials with service name
```

## Best Practices

* Organize secrets hierarchically with descriptive paths
* Use separate paths for different environments (dev, staging, production)
* Limit the number of key-value pairs in a single secret for better management
* Use version history for auditing and rollback capability
* Create specific policies that grant the minimum required access to each secret path
* Regularly rotate sensitive secrets like API keys and passwords
* Use secret delete and undelete for safe secret lifecycle management
* Validate paths are properly formatted and follow naming conventions

## Security Considerations

* Each secret access is authenticated and authorized based on workload identity
* Version history allows for audit trails and secure secret rotation
* Deleted secrets can be recovered if needed
* Secret access is controlled by the `spike policy` permissions system

## Features

* **Store secrets** as key-value pairs at specific paths
* **Retrieve secrets** with full or partial key selection
* **List available secrets** across the system
* **Delete and undelete secret versions** for lifecycle management
* **View secret metadata** to track changes and versioning
* **Path validation** to ensure proper secret organization

## Commands

### `spike secret list`

```bash
spike secret list
```

Lists all available secret paths in the system. Displays paths in a readable 
format.

### `spike secret put`

```bash
spike secret put <path> <key=value>...
```

Stores key-value pairs as a secret at the specified path. Multiple key-value 
pairs can be specified.

#### Examples:

```bash
# Store a single key-value pair
spike secret put secrets/database/creds username=admin

# Store multiple key-value pairs in one command
spike secret put secrets/application/config host=localhost port=8080 debug=true
```

### `spike secret get`

```bash
spike secret get <path> [--version=<version>]
```

Retrieves and displays the key-value pairs stored at the specified secret path. 
By default, returns the current (latest) version, but a specific version can be 
requested.

#### Flags:

| Flag              | Description                                                    |
|-------------------|----------------------------------------------------------------|
| `--version`, `-v` | Specific version to retrieve (default: 0, the current version) |

#### Examples:

```bash
# Get the current version of a secret
spike secret get secrets/database/creds

# Get a specific version of a secret
spike secret get secrets/database/creds --version=2
```

### `spike secret delete`

```bash
spike secret delete <path> [--versions=<versions>]
```

Deletes one or more versions of a secret at the specified path.

#### Flags:

| Flag               | Description                                                                  |
|--------------------|------------------------------------------------------------------------------|
| `--versions`, `-v` | Comma-separated list of versions to delete (default: 0, the current version) |

#### Examples:

```bash
# Delete the current version of a secret
spike secret delete secrets/app/config

# Delete specific versions of a secret
spike secret delete secrets/app/config --versions=1,2,3

# Delete the current version plus specific versions
spike secret delete secrets/app/config --versions=0,1,2
```

### `spike secret undelete`

```bash
spike secret undelete <path> [--versions=<versions>]
```

Restores one or more previously deleted versions of a secret at the specified path.

#### Flags:

| Flag               | Description                                                                   |
|--------------------|-------------------------------------------------------------------------------|
| `--versions`, `-v` | Comma-separated list of versions to restore (default: 0, the current version) |

#### Examples:

```bash
# Restore the current version of a secret
spike secret undelete secrets/app/config

# Restore specific versions of a secret
spike secret undelete secrets/app/config --versions=1,2,3

# Restore the current version plus specific versions
spike secret undelete secrets/app/config --versions=0,1,2
```

### `spike secret metadata get`

```bash
spike secret metadata get <path> [--version=<version>]
```

Retrieves and displays metadata for a secret, including creation time, 
modification time, version history, and other administrative information.

#### Flags:

| Flag              | Description                                                                 |
|-------------------|-----------------------------------------------------------------------------|
| `--version`, `-v` | Specific version to retrieve metadata for (default: 0, the current version) |

#### Examples:

```bash
# Get metadata for the current version of a secret
spike secret metadata get secrets/database/creds

# Get metadata for a specific version of a secret
spike secret metadata get secrets/database/creds --version=2
```

## Path Syntax

Secret paths in **SPIKE** have specific syntax requirements and conventions:

* Paths must match the regex pattern: `^[a-zA-Z0-9._\-/()?+*|[\]{}\\]+$`
* Paths should not have a leading slash
* Using descriptive hierarchical paths is recommended for organization

### Path Examples

```
secrets/app/config                # Application configuration
secrets/database/production/creds # Production database credentials
secrets/certificates/tls          # TLS certificates
```

## Best Practices

* Organize secrets hierarchically with descriptive paths
* Use separate paths for different environments (dev, staging, production)
* Limit the number of key-value pairs in a single secret for better management
* Use version history for auditing and rollback capability
* Create specific policies that grant the minimum required access to each secret path
* Regularly rotate sensitive secrets like API keys and passwords
* Use secret delete and undelete for safe secret lifecycle management
* Validate paths are properly formatted and follow naming conventions

## Security Considerations

* Each secret access is authenticated and authorized based on workload identity
* Version history allows for audit trails and secure secret rotation
* Deleted secrets can be recovered if needed
* Secret access is controlled by the `spike policy` permissions system

----

## `spike` Command Index

{{ toc_commands() }}

----

{{ toc_getting_started() }}

----

{{ toc_top() }}
