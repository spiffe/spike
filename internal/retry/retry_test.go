//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package retry

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/spiffe/spike/internal/retry/mock"
	"github.com/stretchr/testify/require"
)

var errTest = errors.New("test error")

func TestTypedRetrier(t *testing.T) {
	t.Run("successful operation", func(t *testing.T) {
		mockRetrier := &mock.MockRetrier{
			RetryFunc: func(_ context.Context, op func() error) error {
				return op()
			},
		}

		typedRetrier := NewTypedRetrier[string](mockRetrier)
		result, err := typedRetrier.RetryWithBackoff(
			context.Background(),
			func() (string, error) {
				return "success", nil
			},
		)

		require.NoError(t, err)
		require.Equal(t, "success", result)
	})

	t.Run("failed operation", func(t *testing.T) {
		mockRetrier := &mock.MockRetrier{
			RetryFunc: func(_ context.Context, op func() error) error {
				return errTest
			},
		}

		typedRetrier := NewTypedRetrier[string](mockRetrier)
		result, err := typedRetrier.RetryWithBackoff(
			context.Background(),
			func() (string, error) {
				return "", errTest
			},
		)

		require.Equal(t, "", result)
		require.Equal(t, errTest, err)
	})
}

func TestExponentialRetrier(t *testing.T) {
	t.Run("succeeds immediately", func(t *testing.T) {
		retrier := NewExponentialRetrier()
		err := retrier.RetryWithBackoff(
			context.Background(),
			func() error {
				return nil
			},
		)

		require.NoError(t, err)
	})

	t.Run("succeeds after retries", func(t *testing.T) {
		retrier := NewExponentialRetrier()
		attempts := 0

		err := retrier.RetryWithBackoff(
			context.Background(),
			func() error {
				attempts++
				if attempts < 3 {
					return errTest
				}
				return nil
			},
		)

		require.NoError(t, err)
		require.Equal(t, 3, attempts)
	})

	t.Run("respects context cancellation", func(t *testing.T) {
		retrier := NewExponentialRetrier()
		ctx, cancel := context.WithCancel(context.Background())
		attempts := 0

		// Cancel after first attempt
		go func() {
			time.Sleep(10 * time.Millisecond)
			cancel()
		}()

		err := retrier.RetryWithBackoff(
			ctx,
			func() error {
				attempts++
				return errTest
			},
		)

		require.ErrorIs(t, context.Canceled, err)
	})

}

// Example usage in documentation
func ExampleTypedRetrier() {
	// Create a base retrier
	baseRetrier := NewExponentialRetrier()

	// Create a typed retrier for string operations
	stringRetrier := NewTypedRetrier[string](baseRetrier)

	// Use the typed retrier
	result, err := stringRetrier.RetryWithBackoff(
		context.Background(),
		func() (string, error) {
			return "success", nil
		},
	)

	_ = result // Use the result
	_ = err    // Handle error
}
