//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\ Copyright 2024-present SPIKE contributors.
// \\\\\ SPDX-License-Identifier: Apache-2.0

package noop

import (
	"testing"

	sdkErrors "github.com/spiffe/spike-sdk-go/errors"
)

// expectNoError fails the test when a no-op method reports an error.
//
// It is safe to call from goroutines because it uses t.Errorf rather than
// t.Fatalf.
//
// Parameters:
//   - t: The test to fail.
//   - method: The name of the method under test, used in the failure message.
//   - err: The error the method returned; nil is the expected value.
func expectNoError(t *testing.T, method string, err *sdkErrors.SDKError) {
	t.Helper()
	if err != nil {
		t.Errorf("%s should not return an error: %v", method, err)
	}
}
