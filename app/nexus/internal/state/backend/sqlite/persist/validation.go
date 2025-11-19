//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package persist

import (
	"context"

	sdkErrors "github.com/spiffe/spike-sdk-go/errors"
	"github.com/spiffe/spike-sdk-go/log"
)

// validateContext checks if the context is nil and logs a fatal error if so.
func validateContext(ctx context.Context, fName string) {
	if ctx == nil {
		failErr := sdkErrors.ErrNilContext
		log.FatalErr(fName, *failErr)
	}
}
