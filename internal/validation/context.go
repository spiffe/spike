//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package validation

import (
	"context"

	sdkErrors "github.com/spiffe/spike-sdk-go/errors"
	"github.com/spiffe/spike-sdk-go/log"
)

// CheckContext checks if the provided context is nil and terminates the
// program if so.
//
// This function is used to ensure that all operations requiring a context
// receive a valid one. A nil context indicates a programming error that
// should never occur in production, so the function terminates the program
// immediately via log.FatalErr.
//
// Parameters:
//   - ctx: The context to validate
//   - fName: The calling function name for logging purposes
func CheckContext(ctx context.Context, fName string) {
	if ctx == nil {
		failErr := *sdkErrors.ErrNilContext.Clone()
		log.FatalErr(fName, failErr)
	}
}
