//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\ SPDX-License-Identifier: Apache-2.0

package cipher

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/spf13/cobra"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
	apiUrl "github.com/spiffe/spike-sdk-go/api/url"

	"github.com/spiffe/spike/app/spike/internal/stdout"
	"github.com/spiffe/spike/app/spike/internal/trust"
)

func newDecryptCommand(source *workloadapi.X509Source, SPIFFEID string) *cobra.Command {
	var inFile string
	var outFile string

	cmd := &cobra.Command{
		Use:   "decrypt",
		Short: "Decrypt file or stdin via SPIKE Nexus",
		RunE: func(cmd *cobra.Command, args []string) error {
			trust.Authenticate(SPIFFEID)

			var in io.ReadCloser
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

			resp, err := streamToCipherEndpoint(source, apiUrl.NexusCipherDecrypt, in)
			if err != nil {
				return fmt.Errorf("failed to call decrypt endpoint: %w", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				if resp.StatusCode == http.StatusServiceUnavailable {
					stdout.PrintNotReady()
				}
				return fmt.Errorf("decrypt failed: status %d", resp.StatusCode)
			}

			if _, err := io.Copy(out, resp.Body); err != nil {
				return fmt.Errorf("failed to write output: %w", err)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&inFile, "file", "f", "", "Input file (default: stdin)")
	cmd.Flags().StringVarP(&outFile, "out", "o", "", "Output file (default: stdout)")

	return cmd
}
