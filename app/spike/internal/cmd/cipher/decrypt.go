//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\ SPDX-License-Identifier: Apache-2.0

package cipher

import (
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
	sdk "github.com/spiffe/spike-sdk-go/api"

	"github.com/spiffe/spike/app/spike/internal/stdout"
	"github.com/spiffe/spike/app/spike/internal/trust"
)

func newDecryptCommand(source *workloadapi.X509Source, SPIFFEID string) *cobra.Command {
	var inFile string
	var outFile string
	var versionStr string
	var nonceB64 string
	var ciphertextB64 string
	var algorithm string

	cmd := &cobra.Command{
		Use:   "decrypt",
		Short: "Decrypt file or stdin via SPIKE Nexus",
		RunE: func(cmd *cobra.Command, args []string) error {
			trust.AuthenticateForPilot(SPIFFEID)

			jsonMode := versionStr != "" || nonceB64 != "" || ciphertextB64 != ""

			var in io.ReadCloser
			if !jsonMode {
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
			if !jsonMode {
				plaintext, err := api.CipherDecryptStream(in, "application/octet-stream")
				if err != nil {
					if err.Error() == "not ready" {
						stdout.PrintNotReady()
					}
					return fmt.Errorf("failed to call decrypt endpoint: %w", err)
				}
				if _, err := out.Write(plaintext); err != nil {
					return fmt.Errorf("failed to write output: %w", err)
				}
				return nil
			}

			// JSON mode
			v, err := strconv.Atoi(versionStr)
			if err != nil || v < 0 || v > 255 {
				return fmt.Errorf("invalid --version, must be 0-255")
			}
			nonce, err := base64.StdEncoding.DecodeString(nonceB64)
			if err != nil {
				return fmt.Errorf("invalid --nonce base64: %w", err)
			}
			ciphertext, err := base64.StdEncoding.DecodeString(ciphertextB64)
			if err != nil {
				return fmt.Errorf("invalid --ciphertext base64: %w", err)
			}
			plaintext, err := api.CipherDecryptJSON(byte(v), nonce, ciphertext, algorithm)
			if err != nil {
				if err.Error() == "not ready" {
					stdout.PrintNotReady()
				}
				return fmt.Errorf("failed to call decrypt endpoint (json): %w", err)
			}
			if _, err := out.Write(plaintext); err != nil {
				return fmt.Errorf("failed to write plaintext: %w", err)
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&inFile, "file", "f", "", "Input file (default: stdin)")
	cmd.Flags().StringVarP(&outFile, "out", "o", "", "Output file (default: stdout)")
	cmd.Flags().StringVar(&versionStr, "version", "", "Version byte (0-255) for JSON mode")
	cmd.Flags().StringVar(&nonceB64, "nonce", "", "Nonce (base64) for JSON mode")
	cmd.Flags().StringVar(&ciphertextB64, "ciphertext", "", "Ciphertext (base64) for JSON mode")
	cmd.Flags().StringVar(&algorithm, "algorithm", "", "Algorithm hint for JSON mode")

	return cmd
}
