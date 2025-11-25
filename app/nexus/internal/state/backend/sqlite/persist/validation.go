//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package persist

import (
	"context"

	sdkErrors "github.com/spiffe/spike-sdk-go/errors"
	"github.com/spiffe/spike-sdk-go/log"
)

// validateContext checks if the provided context is nil and terminates the
// program if so.
//
// This function is used throughout the persistence layer to ensure that all
// database operations receive a valid context. A nil context indicates a
// programming error that should never occur in production, so the function
// terminates the program immediately via log.FatalErr.
//
// Parameters:
//   - ctx: The context to validate
//   - fName: The calling function name for logging purposes
//
// Returns:
//   - This function does not return if ctx is nil (program terminates)
//   - Returns normally if ctx is valid
func validateContext(ctx context.Context, fName string) {
	if ctx == nil {
		failErr := *sdkErrors.ErrNilContext.Clone()
		log.FatalErr(fName, failErr)
	}
}
