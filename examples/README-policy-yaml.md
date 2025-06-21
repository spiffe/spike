# SPIKE Policy Configuration

This document describes the policy management commands in SPIKE, including YAML file support and command differences.

## Overview

SPIKE provides two commands for managing policies:

1. **`spike policy create`** - Traditional command-line interface (backward compatibility)
2. **`spike policy apply`** - Enhanced command with YAML file support (recommended for new workflows)

Both commands use **upsert semantics** - they will create a new policy if one doesn't exist, or update an existing policy if one with the same name already exists. This makes them safe to use in automation and GitOps workflows.

## Commands

### spike policy create

Traditional command-line interface for policy creation:

```bash
spike policy create \
    --name "web-service-policy" \
    --spiffeid "spiffe://example.org/web-service/*" \
    --path "secrets/web-service/database" \
    --permissions "read,write"
```

**Features:**
- Command-line flags only
- Upsert semantics (create or update)
- Backward compatibility with existing scripts

### spike policy apply (Recommended)

Enhanced command with YAML file support:

```bash
# Using command-line flags
spike policy apply \
    --name "web-service-policy" \
    --spiffeid "spiffe://example.org/web-service/*" \
    --path "secrets/web-service/database" \
    --permissions "read,write"

# Using YAML file (recommended)
spike policy apply --file policy.yaml
```

**Features:**
- YAML file support for GitOps workflows
- Command-line flags (same as create)
- Path normalization (removes trailing slashes)
- Upsert semantics (create or update)
- Better suited for version control

## YAML File Format

### Basic Structure
```yaml
# Policy name - must be unique within the system
name: "web-service-policy"

# SPIFFE ID pattern for workload matching
spiffeid: "spiffe://example.org/web-service/*"

# Path pattern for access control
# Note: Trailing slashes are automatically removed during normalization
path: "secrets/web-service/database"

# List of permissions to grant
permissions:
  - read
  - write
```

### Path Normalization

The `apply` command automatically normalizes paths by removing trailing slashes:

```yaml
# These paths are all normalized to the same value:
path: "secrets/database/production"    # ✓ Normalized form
path: "secrets/database/production/"   # → "secrets/database/production"
path: "secrets/database/production//"  # → "secrets/database/production"

# Special case: root path is preserved
path: "/"  # ✓ Remains as "/"
```

### Realistic Path Examples

```yaml
# Database secrets
name: "database-policy"
spiffeid: "spiffe://example.org/database/*"
path: "secrets/database/production"
permissions: [read]

# Web service configuration
name: "web-service-policy"
spiffeid: "spiffe://example.org/web-service/*"
path: "secrets/web-service/config"
permissions: [read, write]

# Cache credentials
name: "cache-policy"
spiffeid: "spiffe://example.org/cache/*"
path: "secrets/cache/redis/session"
permissions: [read]

# Application environment variables
name: "app-env-policy"
spiffeid: "spiffe://example.org/app/*"
path: "secrets/app/env/production"
permissions: [read, list]
```

### All Available Permissions
```yaml
name: "admin-policy"
spiffeid: "spiffe://example.org/admin/*"
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
spiffeid: "spiffe://example.org/database/*"
path: "secrets/database/production"
permissions: [read, write, list]
```

#### Quoted Values
```yaml
name: "cache-policy"
spiffeid: "spiffe://example.org/cache/*"
path: "secrets/cache/redis"
permissions:
  - "read"
  - "write"
```

## Example Files

The following example files are provided:

- `examples/policy-example.yaml` - Basic policy example
- `examples/test-policies/basic-policy.yaml` - Minimal policy
- `examples/test-policies/admin-policy.yaml` - Full permissions policy
- `examples/test-policies/invalid-permissions.yaml` - Example with invalid permissions (for testing)

## Validation

All policy configurations are validated to ensure:

1. **Required fields**: `name`, `spiffeid`, `path`, and `permissions` must be present
2. **Valid permissions**: Only `read`, `write`, `list`, and `super` are allowed
3. **Valid YAML syntax**: Proper YAML formatting is required (for YAML files)
4. **Non-empty values**: All fields must have non-empty values

## Testing

### Running Tests

To run all policy-related tests:
```bash
go test ./app/spike/internal/cmd/policy -v
```

To run specific test suites:
```bash
# Test YAML file reading functionality
go test ./app/spike/internal/cmd/policy -v -run TestReadPolicyFromFile

# Test command-line flag parsing
go test ./app/spike/internal/cmd/policy -v -run TestGetPolicyFromFlags

# Test policy validation logic
go test ./app/spike/internal/cmd/policy -v -run TestPolicySpecValidation

# Test YAML parsing edge cases
go test ./app/spike/internal/cmd/policy -v -run TestYAMLParsingEdgeCases

# Test path normalization functionality
go test ./app/spike/internal/cmd/policy -v -run TestNormalizePath

# Test apply command functionality
go test ./app/spike/internal/cmd/policy -v -run TestApply
```

### Test Coverage

The test suite covers:

1. **Valid configurations**
   - Basic policy configuration
   - All permission types
   - Different YAML formats (flow sequences, quoted values)
   - Path normalization scenarios

2. **Invalid configurations**
   - Missing required fields (`name`, `spiffeid`, `path`, `permissions`)
   - Empty permission lists
   - Invalid YAML syntax
   - Non-existent files

3. **Command-line flag validation**
   - Valid flag combinations
   - Missing required flags
   - Permission parsing with spaces
   - Empty permission strings

4. **Path normalization**
   - Trailing slash removal
   - Multiple trailing slashes
   - Root path preservation
   - Empty path handling

5. **Edge cases**
   - Multiline YAML values
   - Quoted and unquoted strings
   - Different permission list formats
   - File system errors

## Error Handling

The implementation provides clear error messages for common issues:

- **File not found**: `file policy.yaml does not exist`
- **Invalid YAML**: `failed to parse YAML file policy.yaml: yaml: line 5: found character that cannot start any token`
- **Missing fields**: `policy name is required in YAML file`
- **Invalid permissions**: `invalid permission 'invalid_permission'. Valid permissions are: read, write, list, super`
- **Missing flags**: `required flags are missing: --name, --path (or use --file to read from YAML)`

## GitOps Integration

YAML files can be easily integrated into GitOps workflows:

1. **Store policy YAML files in a Git repository**
   ```bash
   policies/
   ├── web-service-policy.yaml
   ├── database-policy.yaml
   └── admin-policy.yaml
   ```

2. **Use CI/CD pipelines to validate policies before deployment**
   ```bash
   # Validation step in CI
   for policy in policies/*.yaml; do
     spike policy apply --file "$policy" --dry-run
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
5. **Use upsert semantics to safely apply policy changes without worrying about conflicts**

## Migration from API-style Paths

If you have existing policies with API-style paths, consider updating them to use more descriptive paths:

```yaml
# Old API-style paths
path: "/api/v1/secrets/*"
path: "/api/v1/database/*"

# New descriptive paths (recommended)
path: "secrets/database/production"
path: "secrets/web-service/config"
path: "secrets/cache/redis/session"
```

The new path format is:
- More readable and self-documenting
- Better suited for secret management workflows
- Consistent with modern secret management practices
- Automatically normalized (trailing slashes removed) 