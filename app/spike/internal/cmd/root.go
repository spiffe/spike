//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"github.com/spf13/cobra"

	"github.com/spiffe/spike/internal/config"
)

const appName = "SPIKE"

var rootCmd = &cobra.Command{
	Use:   "spike",
	Short: appName + " - Secure your secrets with SPIFFE",
	Long: appName + " v" + config.SpikePilotVersion + `
>> Secure your secrets with SPIFFE: https://spike.ist/ #`,
}
