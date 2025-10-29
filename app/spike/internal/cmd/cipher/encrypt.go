//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\ SPDX-License-Identifier: Apache-2.0

package cipher

import (
	"encoding/base64"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
	sdk "github.com/spiffe/spike-sdk-go/api"

	"github.com/spiffe/spike/app/spike/internal/stdout"
	"github.com/spiffe/spike/app/spike/internal/trust"
)

func newEncryptCommand(source *workloadapi.X509Source, SPIFFEID string) *cobra.Command {
	var inFile string
	var outFile string
	var plaintextB64 string
	var algorithm string

	cmd := &cobra.Command{
		Use:   "encrypt",
		Short: "Encrypt file or stdin via SPIKE Nexus",
		RunE: func(cmd *cobra.Command, args []string) error {
			trust.AuthenticateForPilot(SPIFFEID)

			// Resolve base URL (env or default)
			base := os.Getenv("SPIKE_NEXUS_URL")
			if base == "" {
				base = "https://spire-spike-nexus.spire-server"
			}
			_ = base // base kept for now; endpoint is resolved in helper

			// Prepare input reader (for stream mode)
			var in io.ReadCloser
			if plaintextB64 == "" { // stream mode
				if inFile != "" {
					f, err := os.Open(inFile)
					if err != nil {
						return fmt.Errorf("failed to open input file: %w", err)
					}
					in = f
				} else {
					in = os.Stdin
				}
				defer func() {
					if in != os.Stdin {
						_ = in.Close()
					}
				}()
			}

			// Prepare output writer
			var out io.Writer
			var outCloser io.Closer
			if outFile != "" {
				f, err := os.Create(outFile)
				if err != nil {
					return fmt.Errorf("failed to create output file: %w", err)
				}
				out = f
				outCloser = f
			} else {
				out = os.Stdout
			}
			if outCloser != nil {
				defer outCloser.Close()
			}

			api := sdk.NewWithSource(source)
			if plaintextB64 == "" {
				// Stream mode
				ciphertext, err := api.CipherEncryptStream(in, "application/octet-stream")
				if err != nil {
					if err.Error() == "not ready" {
						stdout.PrintNotReady()
					}
					return fmt.Errorf("failed to call encrypt endpoint: %w", err)
				}
				if _, err := out.Write(ciphertext); err != nil {
					return fmt.Errorf("failed to write output: %w", err)
				}
				return nil
			}

			// JSON mode
			plaintext, err := base64.StdEncoding.DecodeString(plaintextB64)
			if err != nil {
				return fmt.Errorf("invalid --plaintext base64: %w", err)
			}
			ciphertext, err := api.CipherEncryptJSON(plaintext, algorithm)
			if err != nil {
				if err.Error() == "not ready" {
					stdout.PrintNotReady()
				}
				return fmt.Errorf("failed to call encrypt endpoint (json): %w", err)
			}
			if _, err := out.Write(ciphertext); err != nil {
				return fmt.Errorf("failed to write ciphertext: %w", err)
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&inFile, "file", "f", "", "Input file (default: stdin)")
	cmd.Flags().StringVarP(&outFile, "out", "o", "", "Output file (default: stdout)")
	cmd.Flags().StringVar(&plaintextB64, "plaintext", "", "Base64 plaintext for JSON mode; if set, uses JSON API")
	cmd.Flags().StringVar(&algorithm, "algorithm", "", "Algorithm hint for JSON mode")

	return cmd
}
