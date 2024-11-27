//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package base

import (
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/google/uuid"
	"github.com/spiffe/spike-sdk-go/api/entity/data"

	"github.com/spiffe/spike/internal/auth"
)

var (
	ErrPolicyNotFound = errors.New("policy not found")
	ErrPolicyExists   = errors.New("policy already exists")
	ErrInvalidPolicy  = errors.New("invalid policy")
)

func contains(permissions []data.PolicyPermission,
	permission data.PolicyPermission) bool {
	for _, p := range permissions {
		if p == permission {
			return true
		}
	}
	return false
}

func hasAllPermissions(
	haves []data.PolicyPermission,
	wants []data.PolicyPermission,
) bool {
	for _, want := range wants {
		if !contains(haves, want) {
			return false
		}
	}
	return true
}

func CheckAccess(
	spiffeId string, path string, wants []data.PolicyPermission,
) bool {
	if auth.IsPilot(spiffeId) {
		return true
	}

	policies := ListPolicies()
	for _, policy := range policies {
		// Check wildcard pattern first
		if policy.SpiffeIdPattern == "*" && policy.PathPattern == "*" {
			if hasAllPermissions(policy.Permissions, wants) {
				return true
			}
			continue
		}

		// Check specific patterns using pre-compiled regexes

		if policy.SpiffeIdPattern != "*" {
			if !policy.IdRegex.MatchString(spiffeId) {
				continue
			}
		}

		if policy.PathPattern != "*" {
			if !policy.PathRegex.MatchString(path) {
				continue
			}
		}

		if contains(policy.Permissions, data.PermissionSuper) {
			return true
		}

		if hasAllPermissions(policy.Permissions, wants) {
			return true
		}
	}

	return false
}

// CreatePolicy creates a new policy with an auto-generated ID.
func CreatePolicy(policy data.Policy) (data.Policy, error) {
	if policy.Name == "" {
		return data.Policy{}, ErrInvalidPolicy
	}

	// Compile and validate patterns
	if policy.SpiffeIdPattern != "*" {
		idRegex, err := regexp.Compile(policy.SpiffeIdPattern)
		if err != nil {
			return data.Policy{}, fmt.Errorf("%s: %v", "invalid spiffeid pattern", err)
		}
		policy.IdRegex = idRegex
	}

	if policy.PathPattern != "*" {
		pathRegex, err := regexp.Compile(policy.PathPattern)
		if err != nil {
			return data.Policy{}, fmt.Errorf("%s: %v", "invalid path pattern", err)
		}
		policy.PathRegex = pathRegex
	}

	// Generate ID and set creation time
	policy.Id = uuid.New().String()
	if policy.CreatedAt.IsZero() {
		policy.CreatedAt = time.Now()
	}

	policies.Store(policy.Id, policy)
	return policy, nil
}

// GetPolicy retrieves a policy by ID. Returns ErrPolicyNotFound if the policy
// doesn't exist.
func GetPolicy(id string) (data.Policy, error) {
	if value, exists := policies.Load(id); exists {
		return value.(data.Policy), nil
	}
	return data.Policy{}, ErrPolicyNotFound
}

// UpdatePolicy updates an existing policy. Returns ErrPolicyNotFound if
// the policy doesn't exist.
func UpdatePolicy(policy data.Policy) error {
	if policy.Id == "" || policy.Name == "" {
		return ErrInvalidPolicy
	}

	// Check if policy exists
	if _, exists := policies.Load(policy.Id); !exists {
		return ErrPolicyNotFound
	}

	// Preserve original creation timestamp and creator
	original, _ := GetPolicy(policy.Id)
	policy.CreatedAt = original.CreatedAt
	policy.CreatedBy = original.CreatedBy

	policies.Store(policy.Id, policy)
	return nil
}

// DeletePolicy removes a policy by ID. Returns ErrPolicyNotFound if the policy
// doesn't exist.
func DeletePolicy(id string) error {
	if _, exists := policies.Load(id); !exists {
		return ErrPolicyNotFound
	}

	policies.Delete(id)
	return nil
}

// TODO: longer documentation.

// ListPolicies returns all policies as a slice.
func ListPolicies() []data.Policy {
	var result []data.Policy

	policies.Range(func(key, value interface{}) bool {
		result = append(result, value.(data.Policy))
		return true
	})

	return result
}

// ListPoliciesByPath returns all policies that match a given path pattern.
func ListPoliciesByPath(pathPattern string) []data.Policy {
	var result []data.Policy

	policies.Range(func(key, value interface{}) bool {
		policy := value.(data.Policy)
		if policy.PathPattern == pathPattern {
			result = append(result, policy)
		}
		return true
	})

	return result
}

// ListPoliciesBySpiffeId returns all policies that match a given SPIFFE ID pattern.
func ListPoliciesBySpiffeId(spiffeIdPattern string) []data.Policy {
	var result []data.Policy

	policies.Range(func(key, value interface{}) bool {
		policy := value.(data.Policy)
		if policy.SpiffeIdPattern == spiffeIdPattern {
			result = append(result, policy)
		}
		return true
	})

	return result
}
