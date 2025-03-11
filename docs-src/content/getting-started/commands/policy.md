+++
# //    \\ SPIKE: Secure your secrets with SPIFFE.
# //  \\\\\ Copyright 2024-present SPIKE contributors.
# // \\\\\\\ SPDX-License-Identifier: Apache-2.

title = "spike policy"
weight = 3
sort_by = "weight"
+++

# `spike policy`

The `spike policy` command is the main entry point for managing **access
policies** in SPIKE. It allows administrators to define, view, and manage rules
that control access to secrets and resources based on workload identity 
(**SPIFFE ID**) and resource paths.

## Quick Start

```bash
# Create your first policy
spike policy create --name=my-service \
  --path="secrets/app/*" \
  --spiffeid="spiffe://example.org/service/*" \
  --permissions=read

# Verify your policy was created
spike policy list
```

## What are SPIKE Policies?

Policies in **SPIKE** provide a secure and flexible way to control access to 
secrets and resources. Each policy defines:

- **Who** can access resources (via **SPIFFE ID** patterns)
- **What** resources can be accessed (via **path** patterns)
- **How** resources can be accessed (via **permissions**)

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
    - The workload's **SPIFFE ID** against the policy's **SPIFFE ID pattern**
    - The requested **resource path** against the policy's **path pattern**
4. If a match is found, SPIKE checks if the requested operation is allowed by 
  the policy's **permissions**
5. Access is granted only if **ALL** conditions are met

### Why Use Policies?

- **Zero Trust Security**: Access is based on workload identity, not network 
  location
- **Least Privilege**: Grant only the permissions needed for each workload
- **Auditability**: All access is tied to specific policies and identities
- **Flexibility**: Patterns support both exact matching and wildcards
- **Scalability**: Policies work consistently across any deployment size

## Features

- **Create policies** with specific permissions and access patterns
- **List all policies** in human-readable or JSON format
- **Get policy details** by ID or name
- **Delete policies** with confirmation protection
- **Enhanced validation** for permissions and parameters

## Commands

### `spike policy list`

```bash
spike policy list [--format=human|json]
```

Lists all policies in the system. Use `--format=json` for 
machine-readable output.

### `spike policy create`

```bash
spike policy create --name=<name> \
  --path=<path-pattern> \
  --spiffeid=<spiffe-id-pattern> \
  --permissions=<permissions>
```

Creates a new policy with the specified parameters.

#### Permission Types

| Permission | Description                                            |
|------------|--------------------------------------------------------|
| **read**   | Allows reading secrets and resources                   |
| **write**  | Allows creating, updating, and deleting secrets        |
| **list**   | Allows listing resources and directories               |
| **super**  | Full administrative permissions (**use with caution**) |

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
  --path="secrets/web/*" \
  --spiffeid="spiffe://example.org/web/*" \
  --permissions=read,write

# Create a policy with multiple permissions
spike policy create \
  --name=admin-service \
  --path="secrets/*" \
  --spiffeid="spiffe://example.org/admin/*" \
  --permissions=read,write,list

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

**SPIKE** policies support pattern matching for both SPIFFE IDs and 
resource paths:

- `*` matches any sequence of characters within a segment
- Exact matches are also supported for precise control

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
pattern that are provided during policy creation, are used "**AS IS**" to 
create regular expression matchers. These patterns are compiled
into Go's built-in regex engine, ensuring that the matching process strictly
adheres to the patterns defined in the policy, allowing for precise and flexible
access control.

Because of this, a `policy create` operation can define more flexible matching
patterns. However, keeping patterns simple is both more secure and easier to
manage and reason about. Creating a pattern that is too broad or that uses
overly complex regular expressions may lead to unintended consequences and
security risks. **Simplicity** is important to ensure patterns are clear,
predictable, and effective.

When a workload attempts to access a resource, its **SPIFFE ID** and the 
requested resource **path** are matched against these compiled **regular
expressions**. This ensures that both identity and resource patterns follow the
specified rules and allow for flexibility with wildcards or exact matches.

### Path Pattern Examples
```
secrets/*              # All resources in the secrets directory
secrets/database/*     # Only resources in the database subdirectory  
secrets/database/creds # Only the specific creds resource
```

### Path Patterns in SPIKE

Path patterns in **SPIKE** are designed to provide flexibility but also follow
certain conventions for clarity and usability. While the path pattern is
suggested (*but not mandated*) to look like a UNIX-style path for familiarity, 
**SPIKE secret paths DO NOT start with a leading slash**.

This is because **SPIKE paths represent logical key namespaces**, not
hierarchical filesystem paths. They are always relative to the secrets engine
mount point, making the leading slash redundant and potentially confusing.

#### Example:

- **Correct:** `secrets/app/config`
- **Redundant/Confusing:** `/secrets/app/config`

Additionally, although there is currently no restriction on how the path is
formed, it is worth noting that future versions of **SPIKE** may restrict paths 
from having a trailing slash to avoid ambiguity and maintain consistency
in naming practices.

#### Best Practices for Path Patterns:

1. Avoid leading slashes.
2. Avoid trailing slashes to ensure forward compatibility.
3. Use descriptive and meaningful names that reflect the resource's purpose or
   hierarchy.

### SPIFFE ID Pattern Examples
```
spiffe://example.org/*           # All workloads in the example.org trust domain
spiffe://example.org/web/*       # Only web workloads
spiffe://example.org/web/server  # Only the specific web server workload
```

## Best Practices

- Follow the principle of least privilege when assigning permissions
- Use descriptive policy names that reflect their purpose
- Create separate policies for different workload types
- Use specific path patterns rather than overly broad ones
- Regularly audit and review your policies
- Never assign `super` permissions unless absolutely necessary

----

## `spike` Command Index

{{ toc_commands() }}

----

{{ toc_getting_started() }}

----

{{ toc_top() }}