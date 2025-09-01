//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package policy

import (
	"fmt"
	"os"
	"strings"

	"github.com/spiffe/spike-sdk-go/api/entity/data"
	"gopkg.in/yaml.v3"
)

// readPolicyFromFile reads a policy configuration from a YAML file
func readPolicyFromFile(filePath string) (data.PolicySpec, error) {
	var policy data.PolicySpec

	// Check if the file exists:
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return policy, fmt.Errorf("file %s does not exist", filePath)
	}

	// Read file content
	data, err := os.ReadFile(filePath)
	if err != nil {
		return policy, fmt.Errorf("failed to read file %s: %v", filePath, err)
	}

	// Parse YAML
	err = yaml.Unmarshal(data, &policy)
	if err != nil {
		return policy, fmt.Errorf("failed to parse YAML file %s: %v", filePath, err)
	}

	// Validate required fields
	if policy.Name == "" {
		return policy, fmt.Errorf("policy name is required in YAML file")
	}
	if policy.SpiffeIDPattern == "" {
		return policy, fmt.Errorf("spiffeidPattern is required in YAML file")
	}
	if policy.PathPattern == "" {
		return policy, fmt.Errorf("pathPattern is required in YAML file")
	}
	if len(policy.Permissions) == 0 {
		return policy, fmt.Errorf("permissions are required in YAML file")
	}

	return policy, nil
}

// getPolicyFromFlags extracts policy configuration from command line flags
func getPolicyFromFlags(name, SPIFFEIDPattern, pathPattern, permsStr string) (data.PolicySpec, error) {
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
		return policy, fmt.Errorf("required flags are missing: %s (or use --file to read from YAML)", flagList)
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
