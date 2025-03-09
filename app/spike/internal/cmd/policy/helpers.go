package policy

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
	spike "github.com/spiffe/spike-sdk-go/api"
	"github.com/spiffe/spike-sdk-go/api/entity/data"
	"github.com/spiffe/spike/app/spike/internal/stdout"
)

// findPolicyByName searches for a policy with the given name and returns its ID.
// It returns an error if the policy cannot be found or if there's an issue
// with the API call.
//
// Parameters:
//   - api: The SPIKE API client
//   - name: The policy name to search for
//
// Returns:
//   - string: The policy ID if found
//   - error: An error if the policy is not found or there's an API issue
func findPolicyByName(api *spike.Api, name string) (string, error) {
    policies, err := api.ListPolicies()
    if err != nil {
        return "", err
    }

    if policies != nil {
        for _, policy := range *policies {
            if policy.Name == name {
                return policy.Id, nil
            }
        }
    }

    return "", fmt.Errorf("no policy found with name '%s'", name)
}

// addFormatFlag adds a format flag to the given command to allow specifying
// the output format (human or JSON).
//
// Parameters:
//   - cmd: The Cobra command to add the flag to
func addFormatFlag(cmd *cobra.Command) {
    cmd.Flags().String("format", "human", "Output format: 'human' or 'json'")
}

// addNameFlag adds a name flag to the given command to allow specifying
// a policy by name instead of by ID.
//
// Parameters:
//   - cmd: The Cobra command to add the flag to
func addNameFlag(cmd *cobra.Command) {
    cmd.Flags().String("name", "", "Policy name to look up (alternative to policy ID)")
}

// getPolicyID gets the policy ID either from command arguments or the name flag.
// If args contains a policy ID, it returns that. If the name flag is provided,
// it looks up the policy by name and returns its ID. If neither is provided,
// it returns an error.
//
// Parameters:
//   - cmd: The Cobra command containing the flags
//   - args: Command arguments that might contain the policy ID
//   - api: The SPIKE API client
//
// Returns:
//   - string: The policy ID
//   - error: An error if the policy cannot be found or if neither ID nor name is provided
func getPolicyID(cmd *cobra.Command, args []string, api *spike.Api) (string, error) {
    var policyID string
    
    name, _ := cmd.Flags().GetString("name")
    
    if len(args) > 0 {
        policyID = args[0]
    } else if name != "" {
        id, err := findPolicyByName(api, name)
        if err != nil {
            return "", err
        }
        policyID = id
    } else {
        return "", fmt.Errorf("either policy ID as argument or --name flag is required")
    }
    
    return policyID, nil
}

// validatePermissions validates policy permissions from a comma-separated string
// and returns a slice of PolicyPermission values. Only "read", "write", "list",
// and "super" are valid permissions. It returns an error if any permission is invalid
// or if the string contains no valid permissions.
//
// Parameters:
//   - permsStr: Comma-separated string of permissions (e.g., "read,write,list")
//
// Returns:
//   - []data.PolicyPermission: Validated policy permissions
//   - error: An error if any permission is invalid
func validatePermissions(permsStr string) ([]data.PolicyPermission, error) {
    validPerms := map[string]bool{
        "read":  true, 
        "write": true, 
        "list":  true, 
        "super": true,
    }
    
    permissions := []string{}
    for _, p := range strings.Split(permsStr, ",") {
        perm := strings.TrimSpace(p)
        if perm != "" {
            permissions = append(permissions, perm)
        }
    }
    
    perms := make([]data.PolicyPermission, 0, len(permissions))
    for _, perm := range permissions {
        if _, ok := validPerms[perm]; !ok {
            validPermsList := "read, write, list, super"
            return nil, fmt.Errorf("invalid permission '%s'. Valid permissions are: %s", perm, validPermsList)
        }
        perms = append(perms, data.PolicyPermission(perm))
    }
    
    if len(perms) == 0 {
        return nil, fmt.Errorf("no valid permissions specified. Valid permissions are: read, write, list, super")
    }
    
    return perms, nil
}

// checkPolicyNameExists checks if a policy with the given name already exists.
//
// Parameters:
//   - api: The SPIKE API client
//   - name: The policy name to check
//
// Returns:
//   - bool: true if a policy with the name exists, false otherwise
//   - error: An error if there's an issue with the API call
func checkPolicyNameExists(api *spike.Api, name string) (bool, error) {
    policies, err := api.ListPolicies()
    if err != nil {
        return false, err
    }
    
    if policies != nil {
        for _, policy := range *policies {
            if policy.Name == name {
                return true, nil
            }
        }
    }
    
    return false, nil
}

// formatPoliciesOutput formats the output of policies based on the format flag.
// It supports "human" (default) and "json" formats. For human format, it creates
// a readable tabular representation. For JSON format, it marshals the policies
// to indented JSON.
//
// If the format flag is invalid, it returns an error message.
// If the policies list is empty, it returns an appropriate message based on the format.
//
// Parameters:
//   - cmd: The Cobra command containing the format flag
//   - policies: The policies to format
//
// Returns:
//   - string: The formatted output or error message
func formatPoliciesOutput(cmd *cobra.Command, policies *[]data.Policy) string {
    format, _ := cmd.Flags().GetString("format")
    
    // Validate format
    if format != "" && format != "human" && format != "json" {
        return fmt.Sprintf("Error: Invalid format '%s'. Valid formats are: human, json", format)
    }
    
    // Check if policies is nil or empty
    isEmptyList := policies == nil || len(*policies) == 0
    
    if format == "json" {
        if isEmptyList {
            // Return empty array instead of null for empty list in JSON format
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
        return "No policies found"
    }
    
    // Rest of the function remains the same...
    var result strings.Builder
    result.WriteString("POLICIES\n========\n\n")
    
    for _, policy := range *policies {
        result.WriteString(fmt.Sprintf("ID: %s\n", policy.Id))
        result.WriteString(fmt.Sprintf("Name: %s\n", policy.Name))
        result.WriteString(fmt.Sprintf("SPIFFE ID Pattern: %s\n", policy.SpiffeIdPattern))
        result.WriteString(fmt.Sprintf("Path Pattern: %s\n", policy.PathPattern))
        
        perms := make([]string, 0, len(policy.Permissions))
        for _, p := range policy.Permissions {
            perms = append(perms, string(p))
        }
        result.WriteString(fmt.Sprintf("Permissions: %s\n", strings.Join(perms, ", ")))
        result.WriteString(fmt.Sprintf("Created At: %s\n", policy.CreatedAt.Format(time.RFC3339)))
        if policy.CreatedBy != "" {
            result.WriteString(fmt.Sprintf("Created By: %s\n", policy.CreatedBy))
        }
        result.WriteString("--------\n\n")
    }
    
    return result.String()
}

// formatPolicy formats a single policy based on the format flag.
// It converts the policy to a slice and reuses the formatPoliciesOutput function
// for consistent formatting.
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
        return fmt.Sprintf("Error: Invalid format '%s'. Valid formats are: human, json", format)
    }
    
    if policy == nil {
        return "No policy found"
    }
    
    if format == "json" {
        output, err := json.MarshalIndent(policy, "", "  ")
        if err != nil {
            return fmt.Sprintf("Error formatting output: %v", err)
        }
        return string(output)
    }
    
    // Human-readable format for single policy
    var result strings.Builder
    result.WriteString("POLICY DETAILS\n=============\n\n")
    
    result.WriteString(fmt.Sprintf("ID: %s\n", policy.Id))
    result.WriteString(fmt.Sprintf("Name: %s\n", policy.Name))
    result.WriteString(fmt.Sprintf("SPIFFE ID Pattern: %s\n", policy.SpiffeIdPattern))
    result.WriteString(fmt.Sprintf("Path Pattern: %s\n", policy.PathPattern))
    
    perms := make([]string, 0, len(policy.Permissions))
    for _, p := range policy.Permissions {
        perms = append(perms, string(p))
    }
    
    result.WriteString(fmt.Sprintf("Permissions: %s\n", strings.Join(perms, ", ")))
    result.WriteString(fmt.Sprintf("Created At: %s\n", policy.CreatedAt.Format(time.RFC3339)))
    
    if policy.CreatedBy != "" {
        result.WriteString(fmt.Sprintf("Created By: %s\n", policy.CreatedBy))
    }
    
    return result.String()
}

// handleAPIError processes API errors and prints appropriate messages.
// It helps standardize error handling across policy commands.
//
// Parameters:
//   - err: The error returned from an API call
//
// Returns:
//   - bool: true if an error was handled, false if no error
//
// Usage example:
//   policies, err := api.ListPolicies()
//   if handleAPIError(err) {
//       return
//   }
func handleAPIError(err error) bool {
    if err == nil {
        return false
    }
    
    if err.Error() == "not ready" {
        stdout.PrintNotReady()
        return true
    }
    
    if strings.Contains(err.Error(), "unexpected end of JSON") || 
       strings.Contains(err.Error(), "parsing") {
        fmt.Println("Error: Failed to parse API response. The server may be unavailable or returned an invalid response.")
        fmt.Printf("Technical details: %v\n", err)
        return true
    }
    
    fmt.Printf("Error: %v\n", err)
    return true
}