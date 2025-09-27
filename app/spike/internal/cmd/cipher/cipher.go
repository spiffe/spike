//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\ SPDX-License-Identifier: Apache-2.0

package cipher

import (
	"github.com/spf13/cobra"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
)

// NewCipherCommand creates a new Cobra command group for cipher operations.
func NewCipherCommand(source *workloadapi.X509Source, SPIFFEID string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cipher",
		Short: "Encrypt and decrypt data using SPIKE Nexus",
	}

	cmd.AddCommand(newEncryptCommand(source, SPIFFEID))
	cmd.AddCommand(newDecryptCommand(source, SPIFFEID))

	return cmd
}
