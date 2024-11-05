//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"
	"os"

	"github.com/spiffe/go-spiffe/v2/workloadapi"
)

func Initialize(source *workloadapi.X509Source) {
	getCmd := NewGetCommand(source)
	getCmd.Flags().IntP("version", "v", 0, "Specific version to retrieve")
	rootCmd.AddCommand(getCmd)

	deleteCmd := NewDeleteCommand(source)
	deleteCmd.Flags().StringP("versions", "v", "0", "Comma-separated list of versions to delete")
	rootCmd.AddCommand(deleteCmd)

	undeleteCmd := NewUndeleteCommand(source)
	undeleteCmd.Flags().StringP("versions", "v", "0", "Comma-separated list of versions to undelete")
	rootCmd.AddCommand(undeleteCmd)

	initCmd := NewInitCommand(source)
	rootCmd.AddCommand(initCmd)

	putCmd := NewPutCommand(source)
	rootCmd.AddCommand(putCmd)

	listCmd := NewListCommand(source)
	rootCmd.AddCommand(listCmd)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
