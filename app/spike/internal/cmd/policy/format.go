//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package policy

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spiffe/spike-sdk-go/api/entity/data"
	"gopkg.in/yaml.v3"

	"github.com/spiffe/spike/app/spike/internal/cmd/format"
)

// formatPoliciesOutput formats the output of policy list items based on the
// format flag. It supports human/plain, json, and yaml formats. For human
// format, it creates a readable list of policy names (policies are keyed by
// name; there is no separate ID). For JSON/YAML formats, it marshals the
// policies to the appropriate structured format.
//
// If the "policies" list is empty, it returns an appropriate message based on
// the format.
//
// Parameters:
//   - cmd: The Cobra command containing the format flag
//   - policies: The policy list items to format
//
// Returns:
//   - string: The formatted output; empty when an error is returned
//   - error: An error if the format flag is invalid or the policies cannot
//     be marshaled; nil otherwise
func formatPoliciesOutput(
	cmd *cobra.Command, policies *[]data.PolicyListItem,
) (string, error) {
	outputFormat, formatErr := format.GetFormat(cmd)
	if formatErr != nil {
		return "", formatErr
	}

	// Check if "policies" is nil or empty
	isEmptyList := policies == nil || len(*policies) == 0

	switch outputFormat {
	case format.JSON:
		if isEmptyList {
			return "[]", nil
		}
		output, marshalErr := json.MarshalIndent(policies, "", "  ")
		if marshalErr != nil {
			return "", fmt.Errorf("failed to format output: %w", marshalErr)
		}
		return string(output), nil

	case format.YAML:
		if isEmptyList {
			return "[]", nil
		}
		output, marshalErr := yaml.Marshal(policies)
		if marshalErr != nil {
			return "", fmt.Errorf("failed to format output: %w", marshalErr)
		}
		return string(output), nil

	default: // format.Human
		if isEmptyList {
			return "No policies found.", nil
		}

		result := "POLICIES\n========\n\n"

		for _, policy := range *policies {
			result += "Name: " + policy.Name + "\n"
			result += "--------\n\n"
		}

		return result, nil
	}
}

// formatPolicy formats a single policy based on the format flag.
// It supports human/plain, json, and yaml formats.
//
// Parameters:
//   - cmd: The Cobra command containing the format flag
//   - policy: The policy to format; nil yields a "No policy found." message
//
// Returns:
//   - string: The formatted policy; empty when an error is returned
//   - error: An error if the format flag is invalid or the policy cannot be
//     marshaled; nil otherwise
func formatPolicy(
	cmd *cobra.Command, policy *data.Policy,
) (string, error) {
	outputFormat, formatErr := format.GetFormat(cmd)
	if formatErr != nil {
		return "", formatErr
	}

	if policy == nil {
		return "No policy found.", nil
	}

	switch outputFormat {
	case format.JSON:
		output, marshalErr := json.MarshalIndent(policy, "", "  ")
		if marshalErr != nil {
			return "", fmt.Errorf("failed to format output: %w", marshalErr)
		}
		return string(output), nil

	case format.YAML:
		output, marshalErr := yaml.Marshal(policy)
		if marshalErr != nil {
			return "", fmt.Errorf("failed to format output: %w", marshalErr)
		}
		return string(output), nil

	default: // format.Human
		perms := make([]string, 0, len(policy.Permissions))
		for _, p := range policy.Permissions {
			perms = append(perms, string(p))
		}

		result := "POLICY DETAILS\n=============\n\n" +
			"Name: " + policy.Name + "\n" +
			"SPIFFE ID Pattern: " + policy.SPIFFEIDPattern + "\n" +
			"Path Pattern: " + policy.PathPattern + "\n" +
			"Permissions: " + strings.Join(perms, ", ") + "\n" +
			"Created At: " + policy.CreatedAt.Format(time.RFC3339) + "\n"

		if !policy.UpdatedAt.IsZero() {
			result += "Updated At: " +
				policy.UpdatedAt.Format(time.RFC3339) + "\n"
		}

		return result, nil
	}
}
