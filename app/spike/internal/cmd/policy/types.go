//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package policy

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// Spec represents the YAML structure for policy configuration
type Spec struct {
	Name        string   `yaml:"name"`
	SpiffeID    string   `yaml:"spiffeid"`
	Path        string   `yaml:"path"`
	Permissions []string `yaml:"permissions"`
}

// readPolicyFromFile reads a policy configuration from a YAML file
func readPolicyFromFile(filePath string) (Spec, error) {
	var policy Spec

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
	if policy.SpiffeID == "" {
		return policy, fmt.Errorf("spiffeid is required in YAML file")
	}
	if policy.Path == "" {
		return policy, fmt.Errorf("path is required in YAML file")
	}
	if len(policy.Permissions) == 0 {
		return policy, fmt.Errorf("permissions are required in YAML file")
	}

	return policy, nil
}

// getPolicyFromFlags extracts policy configuration from command line flags
func getPolicyFromFlags(name, spiffeIdPattern, pathPattern, permsStr string) (Spec, error) {
	var policy Spec

	// Check if all required flags are provided
	var missingFlags []string
	if name == "" {
		missingFlags = append(missingFlags, "name")
	}
	if pathPattern == "" {
		missingFlags = append(missingFlags, "path")
	}
	if spiffeIdPattern == "" {
		missingFlags = append(missingFlags, "spiffeid")
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
	var permissions []string
	if permsStr != "" {
		for _, perm := range strings.Split(permsStr, ",") {
			perm = strings.TrimSpace(perm)
			if perm != "" {
				permissions = append(permissions, perm)
			}
		}
	}

	policy = Spec{
		Name:        name,
		SpiffeID:    spiffeIdPattern,
		Path:        pathPattern,
		Permissions: permissions,
	}

	return policy, nil
}
