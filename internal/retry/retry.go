//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package retry

import (
	"context"
	"log"
	"time"

	"github.com/cenkalti/backoff/v4"
)

// Retrier handles retry operations with backoff
type Retrier interface {
	// RetryWithBackoff executes an operation with backoff
	RetryWithBackoff(ctx context.Context, op func() error) error
}

// TypedRetrier provides type-safe retry operations
type TypedRetrier[T any] struct {
	retrier Retrier
}

// NewTypedRetrier creates a new TypedRetrier with the given base Retrier
func NewTypedRetrier[T any](r Retrier) *TypedRetrier[T] {
	return &TypedRetrier[T]{retrier: r}
}

// RetryWithBackoff executes a typed operation with backoff
func (r *TypedRetrier[T]) RetryWithBackoff(
	ctx context.Context,
	op func() (T, error),
) (T, error) {
	var result T
	err := r.retrier.RetryWithBackoff(ctx, func() error {
		var err error
		result, err = op()
		return err
	})
	return result, err
}

// ExponentialRetrier implements Retrier using exponential backoff
type ExponentialRetrier struct {
	newBackOff func() backoff.BackOff
}

// NewExponentialRetrier creates a new ExponentialRetrier with default settings
func NewExponentialRetrier() *ExponentialRetrier {
	return &ExponentialRetrier{
		newBackOff: func() backoff.BackOff {
			return backoff.NewExponentialBackOff()
		},
	}
}

// RetryWithBackoff implements the Retrier interface
func (r *ExponentialRetrier) RetryWithBackoff(
	ctx context.Context,
	operation func() error,
) error {
	b := r.newBackOff()
	totalDuration := time.Duration(0)
	return backoff.RetryNotify(
		operation,
		backoff.WithContext(b, ctx),
		func(err error, duration time.Duration) {
			totalDuration += duration
			// log the error, duration and total duration
			log.Printf("Retrying operation after error: %v, duration: %v, total duration: %v", err, duration, totalDuration)
		},
	)
}
