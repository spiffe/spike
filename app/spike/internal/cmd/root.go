//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"github.com/spf13/cobra"

	"github.com/spiffe/spike/internal/config"
)

// appName is the application name used in CLI output and help text.
const appName = "SPIKE"

// rootCmd is the root command for the SPIKE CLI (spike pilot). It serves as
// the entry point for all subcommands including secret management, policy
// management, cipher operations, and operator functions.
//
// The root command itself performs no action but provides the foundation for
// the command hierarchy. Subcommands are registered via the Initialize
// function.
//
// Usage: spike [command] [flags]
var rootCmd = &cobra.Command{
	Use:   "spike",
	Short: appName + " - Secure your secrets with SPIFFE",
	Long: appName + " v" + config.PilotVersion + `
>> Secure your secrets with SPIFFE: https://spike.ist/ #`,
}
