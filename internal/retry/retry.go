//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package retry

import (
	"context"
	"time"

	"github.com/spiffe/spike/internal/log"

	"github.com/cenkalti/backoff/v4"
)

const defaultInitialInterval = 500 * time.Millisecond
const defaultMaxInterval = 3 * time.Second
const defaultMaxElapsedTime = 30 * time.Second

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

// ExponentialRetrierOption is a function type for configuring ExponentialRetrier
type ExponentialRetrierOption func(*backoff.ExponentialBackOff)

// NewExponentialRetrier creates a new ExponentialRetrier with configurable settings
func NewExponentialRetrier(opts ...ExponentialRetrierOption) *ExponentialRetrier {
	if len(opts) == 0 {
		opts = []ExponentialRetrierOption{
			WithInitialInterval(defaultInitialInterval),
			WithMaxInterval(defaultMaxInterval),
			WithMaxElapsedTime(defaultMaxElapsedTime),
		}
	}
	return &ExponentialRetrier{
		newBackOff: func() backoff.BackOff {
			b := backoff.NewExponentialBackOff()
			for _, opt := range opts {
				opt(b)
			}
			return b
		},
	}
}

// WithInitialInterval sets the initial interval between retries
func WithInitialInterval(d time.Duration) ExponentialRetrierOption {
	return func(b *backoff.ExponentialBackOff) {
		b.InitialInterval = d
	}
}

// WithMaxInterval sets the maximum interval between retries
func WithMaxInterval(d time.Duration) ExponentialRetrierOption {
	return func(b *backoff.ExponentialBackOff) {
		b.MaxInterval = d
	}
}

// WithMaxElapsedTime sets the maximum total time for retries
func WithMaxElapsedTime(d time.Duration) ExponentialRetrierOption {
	return func(b *backoff.ExponentialBackOff) {
		b.MaxElapsedTime = d
	}
}

// WithMultiplier sets the multiplier for increasing intervals
func WithMultiplier(m float64) ExponentialRetrierOption {
	return func(b *backoff.ExponentialBackOff) {
		b.Multiplier = m
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
			log.Log().Debug("Retrying operation after error", "error", err.Error(), "duration", duration, "total duration", totalDuration.String())
		},
	)
}
