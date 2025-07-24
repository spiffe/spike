package secret

import (
	"fmt"
	"strings"
	"time"

	"github.com/spiffe/spike-sdk-go/api/entity/data"
)

// formatTime formats a time.Time object into a readable string.
// The format used is "2006-01-02 15:04:05 MST".
func formatTime(t time.Time) string {
	return t.Format("2006-01-02 15:04:05 MST")
}

// printSecretResponse prints secret metadata
func printSecretResponse(response *data.SecretMetadata) {
	printSeparator := func() {
		fmt.Println(strings.Repeat("-", 50))
	}

	hasMetadata := response.Metadata != (data.SecretMetaDataContent{})
	rmd := response.Metadata
	if hasMetadata {
		fmt.Println("\nMetadata:")
		printSeparator()
		fmt.Printf("Current Version : %d\n", rmd.CurrentVersion)
		fmt.Printf("Oldest Version  : %d\n", rmd.OldestVersion)
		fmt.Printf("Created Time    : %s\n", formatTime(rmd.CreatedTime))
		fmt.Printf("Last Updated    : %s\n", formatTime(rmd.UpdatedTime))
		fmt.Printf("Max Versions    : %d\n", rmd.MaxVersions)
		printSeparator()
	}

	if len(response.Versions) > 0 {
		fmt.Println("\nSecret Versions:")
		printSeparator()

		for version, versionData := range response.Versions {
			fmt.Printf("Version %d:\n", version)
			fmt.Printf("  Created: %s\n", formatTime(versionData.CreatedTime))
			if versionData.DeletedTime != nil {
				fmt.Printf("  Deleted: %s\n", formatTime(*versionData.DeletedTime))
			}
			printSeparator()
		}
	}
}
