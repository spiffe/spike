//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package base

import (
	"context"
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
	if spiffeid.IsPilotOperator(peerSPIFFEID) {
		return true
	}

	policies, err := ListPolicies()
	if err != nil {
		log.WarnErr(fName, *sdkErrors.ErrEntityLoadFailed.Clone())
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

		if verifyPermissions(policy.Permissions, wants) {
			return true
		}
	}

	return false
}

// UpsertPolicy creates a new policy or updates an existing one with the same
// name. The function compiles regex patterns, generates a UUID for new policies,
// and sets timestamps appropriately before storing the policy.
//
// This function follows upsert semantics consistent with UpsertSecret:
//   - If no policy with the given name exists, a new policy is created
//   - If a policy with the same name exists, it is updated (ID and CreatedAt
//     are preserved from the existing policy)
//
// Parameters:
//   - policy: The policy to create or update. Must have a non-empty Name field.
//     SPIFFEIDPattern and PathPattern MUST be valid regular expressions.
//
// Returns:
//   - data.Policy: The created or updated policy, including ID and timestamps
//   - *sdkErrors.SDKError: ErrEntityInvalid if the policy name is empty or
//     regex patterns are invalid, ErrEntityLoadFailed or ErrEntitySaveFailed
//     for backend errors
//
// The function performs the following:
//   - Compiles and stores regex patterns for SPIFFEIDPattern and PathPattern
//   - For new policies: generates a UUID, sets CreatedAt and UpdatedAt
//   - For existing policies: preserves ID and CreatedAt, updates UpdatedAt
func UpsertPolicy(policy data.Policy) (data.Policy, *sdkErrors.SDKError) {
	if policy.Name == "" {
		return data.Policy{}, sdkErrors.ErrEntityInvalid.Clone()
	}

	ctx := context.Background()

	// Check for existing policy with the same name
	allPolicies, loadErr := persist.Backend().LoadAllPolicies(ctx)
	if loadErr != nil {
		return data.Policy{}, sdkErrors.ErrEntityLoadFailed.Wrap(loadErr)
	}

	var existingPolicy *data.Policy
	for _, p := range allPolicies {
		if p.Name == policy.Name {
			pCopy := p
			existingPolicy = pCopy
			break
		}
	}

	// Compile and validate patterns
	idRegex, idCompileErr := regexp.Compile(policy.SPIFFEIDPattern)
	if idCompileErr != nil {
		idPatternErr := sdkErrors.ErrEntityInvalid.Clone()
		idPatternErr.Msg = "invalid SPIFFE ID pattern: " + policy.SPIFFEIDPattern +
			" for policy " + policy.Name
		return data.Policy{}, idPatternErr.Wrap(idCompileErr)
	}
	policy.IDRegex = idRegex

	pathRegex, pathCompileErr := regexp.Compile(policy.PathPattern)
	if pathCompileErr != nil {
		pathPatternErr := sdkErrors.ErrEntityInvalid.Clone()
		pathPatternErr.Msg = "invalid path pattern: " + policy.PathPattern +
			" for policy " + policy.Name
		return data.Policy{}, pathPatternErr.Wrap(pathCompileErr)
	}
	policy.PathRegex = pathRegex

	now := time.Now()

	if existingPolicy != nil {
		// Update existing policy: preserve ID and CreatedAt, set UpdatedAt
		policy.ID = existingPolicy.ID
		policy.CreatedAt = existingPolicy.CreatedAt
		policy.UpdatedAt = now
	} else {
		// New policy: generate ID and set creation time
		policy.ID = uuid.New().String()
		if policy.CreatedAt.IsZero() {
			policy.CreatedAt = now
		}
		policy.UpdatedAt = now
	}

	// Store to the backend
	storeErr := persist.Backend().StorePolicy(ctx, policy)
	if storeErr != nil {
		saveErr := sdkErrors.ErrEntitySaveFailed.Clone()
		saveErr.Msg = "failed to store policy " + policy.Name
		return data.Policy{}, saveErr.Wrap(storeErr)
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
//   - *sdkErrors.SDKError: ErrEntityNotFound if no policy exists with the
//     given ID, ErrEntityLoadFailed if loading fails
func GetPolicy(id string) (data.Policy, *sdkErrors.SDKError) {
	ctx := context.Background()

	// Load directly from the backend
	policy, loadErr := persist.Backend().LoadPolicy(ctx, id)
	if loadErr != nil {
		getPolicyErr := sdkErrors.ErrEntityLoadFailed.Clone()
		getPolicyErr.Msg = "failed to load policy with ID " + id
		return data.Policy{}, getPolicyErr.Wrap(loadErr)
	}

	if policy == nil {
		notFoundErr := sdkErrors.ErrEntityNotFound.Clone()
		notFoundErr.Msg = "policy with ID " + id + " not found"
		return data.Policy{}, notFoundErr
	}

	return *policy, nil
}

// DeletePolicy removes a policy from the system by its ID.
//
// Parameters:
//   - id: The unique identifier of the policy to delete
//
// Returns:
//   - *sdkErrors.SDKError: ErrEntityNotFound if no policy exists with the
//     given ID, ErrObjectDeletionFailed if deletion fails, nil on success
func DeletePolicy(id string) *sdkErrors.SDKError {
	ctx := context.Background()

	// Check if the policy exists first (to maintain the same error behavior)
	policy, loadErr := persist.Backend().LoadPolicy(ctx, id)
	if loadErr != nil {
		loadPolicyErr := sdkErrors.ErrEntityLoadFailed.Clone()
		loadPolicyErr.Msg = "failed to load policy with ID " + id
		return loadPolicyErr.Wrap(loadErr)
	}
	if policy == nil {
		notFoundErr := sdkErrors.ErrEntityNotFound.Clone()
		notFoundErr.Msg = "policy with ID " + id + " not found"
		return notFoundErr
	}

	// Delete the policy from the backend
	deleteErr := persist.Backend().DeletePolicy(ctx, id)
	if deleteErr != nil {
		deletePolicyErr := sdkErrors.ErrEntityDeletionFailed.Clone()
		deletePolicyErr.Msg = "failed to delete policy with ID " + id
		return deletePolicyErr.Wrap(deleteErr)
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
//   - *sdkErrors.SDKError: ErrEntityLoadFailed if loading fails, nil on success
func ListPolicies() ([]data.Policy, *sdkErrors.SDKError) {
	ctx := context.Background()

	// Load all policies from the backend
	allPolicies, loadErr := persist.Backend().LoadAllPolicies(ctx)
	if loadErr != nil {
		listPoliciesErr := sdkErrors.ErrEntityLoadFailed.Clone()
		listPoliciesErr.Msg = "failed to load all policies"
		return nil, listPoliciesErr.Wrap(loadErr)
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
//   - *sdkErrors.SDKError: ErrEntityLoadFailed if loading fails, nil on success
func ListPoliciesByPathPattern(
	pathPattern string,
) ([]data.Policy, *sdkErrors.SDKError) {
	ctx := context.Background()

	// Load all policies from the backend
	allPolicies, loadErr := persist.Backend().LoadAllPolicies(ctx)
	if loadErr != nil {
		listByPathErr := sdkErrors.ErrEntityLoadFailed.Clone()
		listByPathErr.Msg = "failed to load policies by pathPattern " + pathPattern
		return nil, listByPathErr.Wrap(loadErr)
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
//   - *sdkErrors.SDKError: ErrEntityLoadFailed if loading fails, nil on success
func ListPoliciesBySPIFFEIDPattern(
	SPIFFEIDPattern string,
) ([]data.Policy, *sdkErrors.SDKError) {
	ctx := context.Background()

	// Load all policies from the backend.
	allPolicies, loadErr := persist.Backend().LoadAllPolicies(ctx)
	if loadErr != nil {
		listByIDErr := sdkErrors.ErrEntityLoadFailed.Clone()
		listByIDErr.Msg = "failed to load policies" +
			" by SPIFFE ID pattern " + SPIFFEIDPattern
		return nil, listByIDErr.Wrap(loadErr)
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
