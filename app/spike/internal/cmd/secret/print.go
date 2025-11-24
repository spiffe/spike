package secret

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spiffe/spike-sdk-go/api/entity/data"
)

// formatTime formats a time.Time value into a human-readable string using
// the format "2006-01-02 15:04:05 MST" (date, time, and timezone).
//
// Parameters:
//   - t: The time value to format
//
// Returns:
//   - string: Formatted time string
//
// Example output: "2024-01-15 14:30:45 UTC"
func formatTime(t time.Time) string {
	return t.Format("2006-01-02 15:04:05 MST")
}

// printSecretResponse formats and prints secret metadata to stdout. The
// function displays secret versioning information including the current
// version, creation time, update time, and per-version details.
//
// The output is formatted in two sections:
//
//  1. Metadata section (if present):
//     - Current version number
//     - Oldest available version
//     - Creation and update timestamps
//     - Maximum versions configured
//
//  2. Versions section:
//     - Per-version creation timestamps
//     - Deletion timestamps (if soft-deleted)
//
// Parameters:
//   - cmd: Cobra command for output
//   - response: Secret metadata containing versioning information
//
// The function uses visual separators (dashes) to improve readability and
// outputs nothing if the metadata is empty.
func printSecretResponse(
	cmd *cobra.Command, response *data.SecretMetadata,
) {
	printSeparator := func() {
		cmd.Println(strings.Repeat("-", 50))
	}

	hasMetadata := response.Metadata != (data.SecretMetaDataContent{})
	rmd := response.Metadata
	if hasMetadata {
		cmd.Println("\nMetadata:")
		printSeparator()
		cmd.Printf("Current Version : %d\n", rmd.CurrentVersion)
		cmd.Printf("Oldest Version  : %d\n", rmd.OldestVersion)
		cmd.Printf("Created Time    : %s\n", formatTime(rmd.CreatedTime))
		cmd.Printf("Last Updated    : %s\n", formatTime(rmd.UpdatedTime))
		cmd.Printf("Max Versions    : %d\n", rmd.MaxVersions)
		printSeparator()
	}

	if len(response.Versions) > 0 {
		cmd.Println("\nSecret Versions:")
		printSeparator()

		for version, versionData := range response.Versions {
			cmd.Printf("Version %d:\n", version)
			cmd.Printf("  Created: %s\n", formatTime(versionData.CreatedTime))
			if versionData.DeletedTime != nil {
				cmd.Printf("  Deleted: %s\n",
					formatTime(*versionData.DeletedTime))
			}
			printSeparator()
		}
	}
}
