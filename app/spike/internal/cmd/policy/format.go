//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package policy

import (
	"encoding/json"
	"fmt"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
	"github.com/spiffe/spike-sdk-go/api/entity/data"
)

// formatPoliciesOutput formats the output of policy list items based on the
// format flag. It supports "human" (default) and "json" formats. For human
// format, it creates a readable tabular representation. For JSON format, it
// marshals the policies to indented JSON.
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
func formatPoliciesOutput(cmd *cobra.Command, policies *[]data.PolicyListItem) string {
	format, _ := cmd.Flags().GetString("format")

	// Validate format
	if format != "" && format != "human" && format != "json" {
		return fmt.Sprintf("Error: Invalid format '%s'."+
			" Valid formats are: human, json", format)
	}

	// Check if "policies" is nil or empty
	isEmptyList := policies == nil || len(*policies) == 0

	if format == "json" {
		if isEmptyList {
			// Return an empty array instead of null for an empty list
			return "[]"
		}

		output, err := json.MarshalIndent(policies, "", "  ")
		if err != nil {
			return fmt.Sprintf("Error formatting output: %v", err)
		}
		return string(output)
	}

	// Default human-readable format
	if isEmptyList {
		return "No policies found."
	}

	// The rest of the function remains the same:
	// Human readable format for multiple policies
	var result strings.Builder
	// Aligns tab-separated columns into a readable table (2-space padding).
	tw := tabwriter.NewWriter(&result, 0, 0, 2, ' ', 0)

	result.WriteString("POLICIES\n========\n\n")

	// Tabwriter header
	fmt.Fprintln(tw, "ID\tNAME")
	for _, policy := range *policies {
		fmt.Fprintf(tw, "%s\t%s\n", policy.ID, policy.Name)
	}

	if err := tw.Flush(); err != nil {
		return fmt.Sprintf("Error: failed to flush tabwriter output: %v\n", err)
	}

	return result.String()
}

// formatPolicy formats a single policy based on the format flag.
// It converts the policy to a slice and reuses the formatPoliciesOutput
// function for consistent formatting.
//
// Parameters:
//   - cmd: The Cobra command containing the format flag
//   - policy: The policy to format
//
// Returns:
//   - string: The formatted policy or error message
func formatPolicy(cmd *cobra.Command, policy *data.Policy) string {
	format, _ := cmd.Flags().GetString("format")

	// Validate format
	if format != "" && format != "human" && format != "json" {
		return fmt.Sprintf("Error: Invalid format '%s'. "+
			"Valid formats are: human, json", format)
	}

	if policy == nil {
		return "No policy found."
	}

	if format == "json" {
		output, err := json.MarshalIndent(policy, "", "  ")
		if err != nil {
			return fmt.Sprintf("Error formatting output: %v", err)
		}
		return string(output)
	}

	// Human-readable format for a single policy:
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
