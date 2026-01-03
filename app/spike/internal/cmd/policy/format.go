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
// format, it creates a readable tabular representation. For JSON/YAML formats,
// it marshals the policies to the appropriate structured format.
//
// If the format flag is invalid, it returns an error message.
// If the "policies" list is empty, it returns an appropriate message based on
// the format.
//
// Parameters:
//   - cmd: The Cobra command containing the format flag
//   - policies: The policy list items to format (contains ID and Name only)
//
// Returns:
//   - string: The formatted output or error message
func formatPoliciesOutput(
	cmd *cobra.Command, policies *[]data.PolicyListItem,
) string {
	outputFormat, formatErr := format.GetFormat(cmd)
	if formatErr != nil {
		return fmt.Sprintf("Error: %v", formatErr)
	}

	// Check if "policies" is nil or empty
	isEmptyList := policies == nil || len(*policies) == 0

	switch outputFormat {
	case format.JSON:
		if isEmptyList {
			return "[]"
		}
		output, marshalErr := json.MarshalIndent(policies, "", "  ")
		if marshalErr != nil {
			return fmt.Sprintf("Error formatting output: %v", marshalErr)
		}
		return string(output)

	case format.YAML:
		if isEmptyList {
			return "[]"
		}
		output, marshalErr := yaml.Marshal(policies)
		if marshalErr != nil {
			return fmt.Sprintf("Error formatting output: %v", marshalErr)
		}
		return string(output)

	default: // format.Human
		if isEmptyList {
			return "No policies found."
		}

		var result strings.Builder
		result.WriteString("POLICIES\n========\n\n")

		for _, policy := range *policies {
			result.WriteString(fmt.Sprintf("ID: %s\n", policy.ID))
			result.WriteString(fmt.Sprintf("Name: %s\n", policy.Name))
			result.WriteString("--------\n\n")
		}

		return result.String()
	}
}

// formatPolicy formats a single policy based on the format flag.
// It supports human/plain, json, and yaml formats.
//
// Parameters:
//   - cmd: The Cobra command containing the format flag
//   - policy: The policy to format
//
// Returns:
//   - string: The formatted policy or error message
func formatPolicy(cmd *cobra.Command, policy *data.Policy) string {
	outputFormat, formatErr := format.GetFormat(cmd)
	if formatErr != nil {
		return fmt.Sprintf("Error: %v", formatErr)
	}

	if policy == nil {
		return "No policy found."
	}

	switch outputFormat {
	case format.JSON:
		output, marshalErr := json.MarshalIndent(policy, "", "  ")
		if marshalErr != nil {
			return fmt.Sprintf("Error formatting output: %v", marshalErr)
		}
		return string(output)

	case format.YAML:
		output, marshalErr := yaml.Marshal(policy)
		if marshalErr != nil {
			return fmt.Sprintf("Error formatting output: %v", marshalErr)
		}
		return string(output)

	default: // format.Human
		var result strings.Builder
		result.WriteString("POLICY DETAILS\n=============\n\n")

		result.WriteString(fmt.Sprintf("ID: %s\n", policy.ID))
		result.WriteString(fmt.Sprintf("Name: %s\n", policy.Name))
		result.WriteString(fmt.Sprintf("SPIFFE ID Pattern: %s\n",
			policy.SPIFFEIDPattern))
		result.WriteString(fmt.Sprintf("Path Pattern: %s\n",
			policy.PathPattern))

		perms := make([]string, 0, len(policy.Permissions))
		for _, p := range policy.Permissions {
			perms = append(perms, string(p))
		}

		result.WriteString(fmt.Sprintf("Permissions: %s\n",
			strings.Join(perms, ", ")))
		result.WriteString(fmt.Sprintf("Created At: %s\n",
			policy.CreatedAt.Format(time.RFC3339)))

		if !policy.UpdatedAt.IsZero() {
			result.WriteString(fmt.Sprintf("Updated At: %s\n",
				policy.UpdatedAt.Format(time.RFC3339)))
		}

		return result.String()
	}
}
