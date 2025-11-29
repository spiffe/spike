//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package policy

import (
	"os"
	"strings"

	"github.com/spiffe/spike-sdk-go/api/entity/data"
	sdkErrors "github.com/spiffe/spike-sdk-go/errors"
	"gopkg.in/yaml.v3"
)

// readPolicyFromFile reads and parses a policy configuration from a YAML
// file. The function validates that the file exists, parses the YAML content,
// and ensures all required fields are present.
//
// The YAML file must contain the following required fields:
//   - name: Policy name
//   - spiffeidPattern: Regular expression pattern for SPIFFE IDs
//   - pathPattern: Regular expression pattern for resource paths
//   - permissions: List of permissions to grant
//
// Parameters:
//   - filePath: Path to the YAML file containing the policy specification
//
// Returns:
//   - data.PolicySpec: Parsed policy specification
//   - *sdkErrors.SDKError: File reading, parsing, or validation errors
//
// Example YAML format:
//
//	name: my-policy
//	spiffeidPattern: "^spiffe://example\\.org/.*$"
//	pathPattern: "^secrets/.*$"
//	permissions:
//	  - read
//	  - write
func readPolicyFromFile(
	filePath string,
) (data.PolicySpec, *sdkErrors.SDKError) {
	var policy data.PolicySpec

	// Check if the file exists:
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		failErr := sdkErrors.ErrFSFileOpenFailed.Clone()
		failErr.Msg = "file " + filePath + " does not exist"
		return policy, failErr
	}

	// Read file content
	df, err := os.ReadFile(filePath)
	if err != nil {
		failErr := sdkErrors.ErrFSStreamReadFailed.Wrap(err)
		failErr.Msg = "failed to read file " + filePath
		return policy, failErr
	}

	// Parse YAML
	err = yaml.Unmarshal(df, &policy)
	if err != nil {
		failErr := sdkErrors.ErrDataUnmarshalFailure.Wrap(err)
		failErr.Msg = "failed to parse YAML file " + filePath
		return policy, failErr
	}

	// Validate required fields
	if policy.Name == "" {
		failErr := sdkErrors.ErrDataInvalidInput.Clone()
		failErr.Msg = "policy name is required in YAML file"
		return policy, failErr
	}
	if policy.SpiffeIDPattern == "" {
		failErr := sdkErrors.ErrDataInvalidInput.Clone()
		failErr.Msg = "spiffeidPattern is required in YAML file"
		return policy, failErr
	}
	if policy.PathPattern == "" {
		failErr := sdkErrors.ErrDataInvalidInput.Clone()
		failErr.Msg = "pathPattern is required in YAML file"
		return policy, failErr
	}
	if len(policy.Permissions) == 0 {
		failErr := sdkErrors.ErrDataInvalidInput.Clone()
		failErr.Msg = "permissions are required in YAML file"
		return policy, failErr
	}

	return policy, nil
}

// getPolicyFromFlags constructs a policy specification from command-line
// flag values. The function validates that all required flags are provided
// and parses the comma-separated permissions string into a slice.
//
// All parameters are required. If any parameter is empty, the function
// returns an error listing all missing flags.
//
// Parameters:
//   - name: Policy name (required)
//   - SPIFFEIDPattern: Regular expression pattern for SPIFFE IDs (required)
//   - pathPattern: Regular expression pattern for resource paths (required)
//   - permsStr: Comma-separated list of permissions (e.g., "read,write")
//
// Returns:
//   - data.PolicySpec: Constructed policy specification
//   - *sdkErrors.SDKError: Validation errors if required flags are missing
//
// Example usage:
//
//	policy, err := getPolicyFromFlags(
//	    "my-policy",
//	    "^spiffe://example\\.org/.*$",
//	    "^secrets/.*$",
//	    "read,write,delete",
//	)
func getPolicyFromFlags(
	name, SPIFFEIDPattern, pathPattern, permsStr string,
) (data.PolicySpec, *sdkErrors.SDKError) {
	var policy data.PolicySpec

	// Check if all required flags are provided
	var missingFlags []string
	if name == "" {
		missingFlags = append(missingFlags, "name")
	}
	if pathPattern == "" {
		missingFlags = append(missingFlags, "path-pattern")
	}
	if SPIFFEIDPattern == "" {
		missingFlags = append(missingFlags, "spiffeid-pattern")
	}
	if permsStr == "" {
		missingFlags = append(missingFlags, "permissions")
	}

	if len(missingFlags) > 0 {
		flagList := ""
		for i, flag := range missingFlags {
			if i > 0 {
				flagList += ", "
			}
			flagList += "--" + flag
		}
		failErr := sdkErrors.ErrDataInvalidInput.Clone()
		failErr.Msg = "required flags are missing: " + flagList +
			" (or use --file to read from YAML)"
		return policy, failErr
	}

	// Convert comma-separated permissions to slice
	var permissions []data.PolicyPermission
	if permsStr != "" {
		for _, perm := range strings.Split(permsStr, ",") {
			perm = strings.TrimSpace(perm)
			if perm != "" {
				permissions = append(permissions, data.PolicyPermission(perm))
			}
		}
	}

	policy = data.PolicySpec{
		Name:            name,
		SpiffeIDPattern: SPIFFEIDPattern,
		PathPattern:     pathPattern,
		Permissions:     permissions,
	}

	return policy, nil
}
