//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\ SPDX-License-Identifier: Apache-2.0

package cipher

import (
	"github.com/spf13/cobra"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
)

// NewCipherCommand creates a new top-level command for cryptographic
// operations. It acts as a parent for all cipher-related subcommands:
// encrypt and decrypt.
//
// The cipher commands provide encryption and decryption capabilities through
// SPIKE Nexus, allowing workloads to protect sensitive data in transit or at
// rest. Operations can work with files, stdin/stdout, or base64-encoded
// strings.
//
// Parameters:
//   - source: SPIFFE X.509 SVID source for authentication
//   - SPIFFEID: The SPIFFE ID to authenticate with
//
// Returns:
//   - *cobra.Command: Configured top-level Cobra command for cipher operations
//
// Available subcommands:
//   - encrypt: Encrypt data via SPIKE Nexus
//   - decrypt: Decrypt data via SPIKE Nexus
//
// Example usage:
//
//	spike cipher encrypt --in secret.txt --out secret.enc
//	spike cipher decrypt --in secret.enc --out secret.txt
//	echo "sensitive data" | spike cipher encrypt | spike cipher decrypt
//
// Each subcommand supports multiple input/output modes including files,
// stdin/stdout streams, and base64-encoded strings. See the individual
// command documentation for details.
func NewCipherCommand(
	source *workloadapi.X509Source, SPIFFEID string,
) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cipher",
		Short: "Encrypt and decrypt data using SPIKE Nexus",
	}

	cmd.AddCommand(newEncryptCommand(source, SPIFFEID))
	cmd.AddCommand(newDecryptCommand(source, SPIFFEID))

	return cmd
}
