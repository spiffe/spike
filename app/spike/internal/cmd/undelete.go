//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
)

func NewUndeleteCommand(source *workloadapi.X509Source) *cobra.Command {
	var undeleteCmd = &cobra.Command{
		Use:   "undelete <path>",
		Short: "Undelete secrets at the specified path",
		Long: `Undelete secrets at the specified path. 
Specify versions using -v or --versions flag with comma-separated values.
Version 0 refers to the current/latest version.
If no version is specified, defaults to undeleting the current version.

Examples:
  spike undelete secret/ella           # Undeletes current version
  spike undelete secret/ella -v all    # Undeletes all versions
  spike undelete secret/ella -v 1,2,3  # Undeletes specific versions
  spike undelete secret/ella -v 0,1,2  # Undeletes current version plus versions 1 and 2`,
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			path := args[0]
			versions, _ := cmd.Flags().GetString("versions")

			fmt.Println("###### NEEDS TO BE IMPLEMENTED ######")

			if versions == "all" {
				fmt.Printf("Undeleting all versions at path %s\n", path)
				return
			}

			if versions == "" || versions == "0" {
				fmt.Printf("Undeleting current version at path %s\n", path)
				return
			}

			// Parse and validate versions
			versionList := strings.Split(versions, ",")
			for _, v := range versionList {
				version, err := strconv.Atoi(strings.TrimSpace(v))
				if err != nil {
					fmt.Printf("Error: invalid version number: %s\n", v)
					return
				}
				if version < 0 {
					fmt.Printf("Error: version numbers cannot be negative: %s\n", v)
					return
				}
			}

			if strings.Contains(versions, "0") {
				fmt.Printf("Undeleting current version and versions %s at path %s\n",
					strings.Replace(versions, "0,", "", 1), path)
			} else {
				fmt.Printf("Undeleting versions %s at path %s\n", versions, path)
			}
		},
	}

	return undeleteCmd
}
