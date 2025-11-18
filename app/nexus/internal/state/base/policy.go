//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package base

import (
	"context"
	"errors"
	"regexp"
	"time"

	"github.com/google/uuid"
	"github.com/spiffe/spike-sdk-go/api/entity/data"
	sdkErrors "github.com/spiffe/spike-sdk-go/errors"
	"github.com/spiffe/spike-sdk-go/log"
	"github.com/spiffe/spike-sdk-go/spiffeid"

	"github.com/spiffe/spike/app/nexus/internal/state/persist"
)

// CheckAccess determines if a given SPIFFE ID has the required permissions for
// a specific path. It first checks if the ID belongs to SPIKE Pilot (which has
// unrestricted access), then evaluates against all defined policies. Policies
// are checked in order, with wildcard patterns evaluated first, followed by
// specific pattern matching using regular expressions.
//
// Parameters:
//   - spiffeId: The SPIFFE ID of the requestor
//   - path: The resource path being accessed
//   - wants: Slice of permissions being requested
//
// Returns:
//   - bool: true if access is granted, false otherwise
//
// The function grants access if any of these conditions are met:
//  1. The requestor is a SPIKE Pilot instance.
//  2. A matching policy has the super permission
//  3. A matching policy contains all requested permissions
//
// A policy matches when:
//
//	Its SPIFFE ID pattern matches the requestor's ID, and its path pattern
//	matches the requested path.
func CheckAccess(
	peerSPIFFEID string, path string, wants []data.PolicyPermission,
) bool {
	const fName = "CheckAccess"
	// Role:SpikePilot can always manage secrets and policies,
	// and can call encryption and decryption API endpoints.
	if spiffeid.IsPilot(peerSPIFFEID) {
		return true
	}

	policies, err := ListPolicies()
	if err != nil {
		log.Log().Warn(
			fName,
			"message", sdkErrors.ErrCodeResultSetFailedToLoad,
			"err", err.Error(),
		)
		return false
	}

	for _, policy := range policies {
		// Check specific patterns using pre-compiled regexes

		if !policy.IDRegex.MatchString(peerSPIFFEID) {
			continue
		}

		if !policy.PathRegex.MatchString(path) {
			continue
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

// CreatePolicy creates a new policy in the system after validating and
// preparing it. The function compiles regex patterns, generates a UUID, and
// sets the creation timestamp before storing the policy.
//
// Parameters:
//   - policy: The policy to create. Must have a non-empty Name field.
//     SpiffeIdPattern and PathPattern MUST be valid regular expressions.
//
// Returns:
//   - data.Policy: The created policy, including generated ID and timestamps
//   - error: ErrInvalidPolicy if policy name is empty, or regex compilation
//     errors for invalid patterns
//
// The function performs the following modifications to the input policy:
//   - Compiles and stores regex patterns for non-wildcard SpiffeIdPattern
//     and PathPattern
//   - Generates and sets a new UUID as the policy ID
//   - Sets CreatedAt to current time if not already set
func CreatePolicy(policy data.Policy) (data.Policy, error) {
	const fName = "CreatePolicy"

	if policy.Name == "" {
		return data.Policy{}, sdkErrors.ErrPolicyInvalid
	}

	ctx := context.Background()

	// Check for duplicate policy name
	allPolicies, err := persist.Backend().LoadAllPolicies(ctx)
	if err != nil {
		failErr := sdkErrors.ErrDataLoadFailed
		return data.Policy{}, errors.Join(failErr, err)
	}

	for _, existingPolicy := range allPolicies {
		if existingPolicy.Name == policy.Name {
			return data.Policy{}, sdkErrors.ErrPolicyExists
		}
	}

	// Compile and validate patterns
	idRegex, err := regexp.Compile(policy.SPIFFEIDPattern)
	if err != nil {
		failMsg := sdkErrors.InvalidFor(
			"SPIFFEID pattern", "policy", policy.SPIFFEIDPattern,
		)
		log.Log().Warn(fName, "message", failMsg)
		return data.Policy{}, errors.Join(sdkErrors.ErrPolicyInvalid, err)
	}
	policy.IDRegex = idRegex

	pathRegex, err := regexp.Compile(policy.PathPattern)
	if err != nil {
		failMsg := sdkErrors.InvalidFor(
			"path pattern", "policy", policy.PathPattern,
		)
		log.Log().Warn(fName, "message", failMsg)
		return data.Policy{}, errors.Join(sdkErrors.ErrPolicyInvalid, err)
	}
	policy.PathRegex = pathRegex

	// Generate ID and set creation time
	policy.ID = uuid.New().String()
	if policy.CreatedAt.IsZero() {
		policy.CreatedAt = time.Now()
	}

	// Store directly to the backend
	err = persist.Backend().StorePolicy(ctx, policy)
	if err != nil {
		failErr := sdkErrors.ErrDataSaveFailed
		return data.Policy{}, errors.Join(failErr, err)
	}

	return policy, nil
}

// GetPolicy retrieves a policy by its ID from the policy store.
//
// Parameters:
//   - id: The unique identifier of the policy to retrieve
//
// Returns:
//   - data.Policy: The retrieved policy if found
//   - error: ErrPolicyNotFound if no policy exists with the given ID.
func GetPolicy(id string) (data.Policy, error) {
	ctx := context.Background()

	// Load directly from the backend
	policy, err := persist.Backend().LoadPolicy(ctx, id)
	if err != nil {
		failErr := sdkErrors.ErrDataLoadFailed
		return data.Policy{}, errors.Join(failErr, err)
	}

	if policy == nil {
		return data.Policy{}, sdkErrors.ErrPolicyNotFound
	}

	return *policy, nil
}

// DeletePolicy removes a policy from the system by its ID.
//
// Parameters:
//   - id: The unique identifier of the policy to delete
//
// Returns:
//   - error: ErrPolicyNotFound if no policy exists with the given ID,
//     nil if the deletion was successful
func DeletePolicy(id string) error {
	ctx := context.Background()

	// Check if the policy exists first (to maintain the same error behavior)
	policy, err := persist.Backend().LoadPolicy(ctx, id)
	if err != nil {
		failErr := sdkErrors.ErrDataLoadFailed
		return errors.Join(failErr, err)
	}
	if policy == nil {
		return sdkErrors.ErrPolicyNotFound
	}

	// Delete the policy from the backend
	err = persist.Backend().DeletePolicy(ctx, id)
	if err != nil {
		failErr := sdkErrors.ErrDeletionFailed
		return errors.Join(failErr, err)
	}

	return nil
}

// ListPolicies retrieves all policies from the policy store.
// It iterates through the concurrent map of policies and returns them as a
// slice.
//
// Returns:
//   - []data.Policy: A slice containing all existing policies. Returns an empty
//     slice if no policies exist. The order of policies in the returned slice
//     is non-deterministic due to the concurrent nature of the underlying
//     store.
func ListPolicies() ([]data.Policy, error) {
	ctx := context.Background()

	// Load all policies from the backend
	allPolicies, err := persist.Backend().LoadAllPolicies(ctx)
	if err != nil {
		failErr := sdkErrors.ErrDataLoadFailed
		return nil, errors.Join(failErr, err)
	}

	// Convert map to slice
	result := make([]data.Policy, 0, len(allPolicies))
	for _, policy := range allPolicies {
		if policy != nil {
			result = append(result, *policy)
		}
	}

	return result, nil
}

// ListPoliciesByPathPattern returns all policies that match a specific
// pathPattern pattern. It filters the policy store and returns only policies
// where PathPattern exactly matches the provided pattern string.
//
// Parameters:
//   - pathPattern: The exact pathPattern pattern to match against policies
//
// Returns:
//   - []data.Policy: A slice of policies with matching PathPattern. Returns an
//     empty slice if no policies match. The order of policies in the returned
//     slice is non-deterministic due to the concurrent nature of the underlying
//     store.
func ListPoliciesByPathPattern(pathPattern string) ([]data.Policy, error) {
	ctx := context.Background()

	// Load all policies from the backend
	allPolicies, err := persist.Backend().LoadAllPolicies(ctx)
	if err != nil {
		failErr := sdkErrors.ErrDataLoadFailed
		return nil, errors.Join(failErr, err)
	}

	// Filter by pathPattern pattern
	var result []data.Policy
	for _, policy := range allPolicies {
		if policy != nil && policy.PathPattern == pathPattern {
			result = append(result, *policy)
		}
	}

	return result, nil
}

// ListPoliciesBySPIFFEIDPattern returns all policies that match a specific
// SPIFFE ID pattern. It filters the policy store and returns only policies
// where SpiffeIdPattern exactly matches the provided pattern string.
//
// Parameters:
//   - spiffeIdPattern: The exact SPIFFE ID pattern to match against policies
//
// Returns:
//   - []data.Policy: A slice of policies with matching SpiffeIdPattern. Returns
//     an empty slice if no policies match. The order of policies in the
//     returned slice is non-deterministic due to the concurrent nature of the
//     underlying store.
func ListPoliciesBySPIFFEIDPattern(
	SPIFFEIDPattern string,
) ([]data.Policy, error) {
	ctx := context.Background()

	// Load all policies from the backend.
	allPolicies, err := persist.Backend().LoadAllPolicies(ctx)
	if err != nil {
		failErr := sdkErrors.ErrDataLoadFailed
		return nil, errors.Join(failErr, err)
	}

	// Filter by SPIFFE ID pattern
	var result []data.Policy
	for _, policy := range allPolicies {
		if policy != nil && policy.SPIFFEIDPattern == SPIFFEIDPattern {
			result = append(result, *policy)
		}
	}

	return result, nil
}
