//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package cipher

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spiffe/spike-sdk-go/config/env"
	"github.com/spiffe/spike-sdk-go/spiffeid"
)

// pilotSPIFFEID returns a SPIFFE ID that passes the Pilot operator gate for
// the trust root configured in the test environment.
//
// Returns:
//   - string: The Pilot SPIFFE ID for the first configured Pilot trust root
func pilotSPIFFEID() string {
	trustRoots := env.TrustRootFromEnv(env.TrustRootPilot)
	first := strings.TrimSpace(strings.Split(trustRoots, ",")[0])
	return spiffeid.Pilot(first)
}

// newSilencedRoot mirrors the Pilot root command: handlers return errors,
// and Cobra's own error and usage output is silenced so that the root
// command can print the error once and exit non-zero.
//
// Parameters:
//   - sub: The command under test, attached as a child of the root
//
// Returns:
//   - *cobra.Command: A "spike" root command with SilenceErrors and
//     SilenceUsage set and sub registered as its subcommand
func newSilencedRoot(sub *cobra.Command) *cobra.Command {
	root := &cobra.Command{
		Use:           "spike",
		SilenceErrors: true,
		SilenceUsage:  true,
	}
	root.AddCommand(sub)
	return root
}

// isSyntaxError reports whether err wraps the error encoding/json returns
// for input that is not JSON at all.
//
// Parameters:
//   - err: The error to inspect
//
// Returns:
//   - bool: true if err wraps a *json.SyntaxError
func isSyntaxError(err error) bool {
	var syntaxErr *json.SyntaxError
	return errors.As(err, &syntaxErr)
}

// typeErrorOn returns a check that err wraps the encoding/json type error
// for the named field, so a type error from a different field fails it.
//
// Parameters:
//   - field: The JSON field the type error must name
//
// Returns:
//   - func(error) bool: Reports whether its argument wraps a
//     *json.UnmarshalTypeError for field
func typeErrorOn(field string) func(error) bool {
	return func(err error) bool {
		var typeErr *json.UnmarshalTypeError
		return errors.As(err, &typeErr) && typeErr.Field == field
	}
}

// isBase64Error reports whether err wraps the error encoding/json returns
// when a byte field is not valid base64.
//
// Parameters:
//   - err: The error to inspect
//
// Returns:
//   - bool: true if err wraps a base64.CorruptInputError
func isBase64Error(err error) bool {
	var base64Err base64.CorruptInputError
	return errors.As(err, &base64Err)
}
