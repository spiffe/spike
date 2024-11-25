//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package mock

import (
	"context"
)

// MockRetrier implements Retrier for testing
type MockRetrier struct {
	RetryFunc func(context.Context, func() error) error
}

// RetryWithBackoff implements the Retrier interface
func (m *MockRetrier) RetryWithBackoff(
	ctx context.Context,
	operation func() error,
) error {
	if m.RetryFunc != nil {
		return m.RetryFunc(ctx, operation)
	}
	return nil
}
