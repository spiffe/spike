//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package cipher

import (
	"github.com/spf13/cobra"
	sdkErrors "github.com/spiffe/spike-sdk-go/errors"

	"github.com/spiffe/spike/app/spike/internal/stdout"
)

// cipherCommandPath is a stand-in command hierarchy ("spike cipher") that
// lets stdout.APIError apply the cipher-specific error mapping without the
// implementation functions having to carry the live Cobra command around.
var cipherCommandPath = func() *cobra.Command {
	root := &cobra.Command{Use: "spike"}
	group := &cobra.Command{Use: "cipher"}
	root.AddCommand(group)
	return group
}()

// cipherAPIError maps an SDK error from a cipher API call to the error the
// command returns.
//
// Parameters:
//   - err: The SDK error returned by the cipher API call
//
// Returns:
//   - error: The user-facing error produced by stdout.APIError under the
//     cipher command group; nil when err is nil
func cipherAPIError(err *sdkErrors.SDKError) error {
	return stdout.APIError(cipherCommandPath, err)
}
