//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

// Package policy implements SPIKE CLI commands for managing access control
// policies.
//
// Policies define which workloads can access which secrets based on SPIFFE ID
// patterns, path patterns, and permissions. Each policy grants a set of
// permissions to workloads whose SPIFFE IDs match the policy's SPIFFE ID
// pattern, for secrets whose paths match the policy's path pattern.
//
// Available commands:
//
//   - create: Create a new policy with the specified name, patterns, and
//     permissions. Fails if a policy with the same name already exists.
//   - apply: Create or update a policy (upsert semantics). Accepts flags or
//     a YAML file for configuration.
//   - list: List all policies, with optional filtering by name pattern.
//   - get: Retrieve a specific policy by ID or name.
//   - delete: Remove a policy by ID or name.
//
// Pattern matching:
//
// Both spiffeid-pattern and path-pattern use regular expressions (not globs).
// For example:
//
//	--spiffeid-pattern "^spiffe://example\.org/web-service/.*$"
//	--path-pattern "^tenants/acme/creds/.*$"
//
// Available permissions:
//
//   - read: Read secret values
//   - write: Create, update, or delete secrets
//   - list: List secret paths
//   - super: Administrative access (grants all permissions)
//
// Example usage:
//
//	spike policy create \
//	    --name "web-service-policy" \
//	    --spiffeid-pattern "^spiffe://example\.org/web/.*$" \
//	    --path-pattern "^secrets/web/.*$" \
//	    --permissions read,list
//
//	spike policy apply --file policy.yaml
//	spike policy list
//	spike policy get --name web-service-policy
//	spike policy delete --name web-service-policy
//
// See https://spike.ist/usage/commands/ for detailed CLI documentation.
package policy
