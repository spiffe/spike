//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package retry

import (
	"context"
	"time"

	"github.com/cenkalti/backoff/v4"
)

const (
	defaultInitialInterval = 500 * time.Millisecond
	defaultMaxInterval     = 3 * time.Second
	defaultMaxElapsedTime  = 30 * time.Second
	defaultMultiplier      = 2.0
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

// NotifyFn is a callback function type for retry notifications
type NotifyFn func(err error, duration, totalDuration time.Duration)

// ExponentialRetrier implements Retrier using exponential backoff
type ExponentialRetrier struct {
	newBackOff func() backoff.BackOff
	notify     NotifyFn
}

// RetrierOption is a function type for configuring ExponentialRetrier
type RetrierOption func(*ExponentialRetrier)

// BackOffOption is a function type for configuring ExponentialBackOff
type BackOffOption func(*backoff.ExponentialBackOff)

// NewExponentialRetrier creates a new ExponentialRetrier with configurable settings
func NewExponentialRetrier(opts ...RetrierOption) *ExponentialRetrier {
	b := backoff.NewExponentialBackOff()
	b.InitialInterval = defaultInitialInterval
	b.MaxInterval = defaultMaxInterval
	b.MaxElapsedTime = defaultMaxElapsedTime
	b.Multiplier = defaultMultiplier

	r := &ExponentialRetrier{
		newBackOff: func() backoff.BackOff {
			return b
		},
	}

	for _, opt := range opts {
		opt(r)
	}

	return r
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
			if r.notify != nil {
				r.notify(err, duration, totalDuration)
			}
		},
	)
}

// WithBackOffOptions configures the backoff settings
func WithBackOffOptions(opts ...BackOffOption) RetrierOption {
	return func(r *ExponentialRetrier) {
		b := r.newBackOff().(*backoff.ExponentialBackOff)
		for _, opt := range opts {
			opt(b)
		}
	}
}

// WithInitialInterval sets the initial interval between retries
func WithInitialInterval(d time.Duration) BackOffOption {
	return func(b *backoff.ExponentialBackOff) {
		b.InitialInterval = d
	}
}

// WithMaxInterval sets the maximum interval between retries
func WithMaxInterval(d time.Duration) BackOffOption {
	return func(b *backoff.ExponentialBackOff) {
		b.MaxInterval = d
	}
}

// WithMaxElapsedTime sets the maximum total time for retries
func WithMaxElapsedTime(d time.Duration) BackOffOption {
	return func(b *backoff.ExponentialBackOff) {
		b.MaxElapsedTime = d
	}
}

// WithMultiplier sets the multiplier for increasing intervals
func WithMultiplier(m float64) BackOffOption {
	return func(b *backoff.ExponentialBackOff) {
		b.Multiplier = m
	}
}

// WithNotify is an option to set the notification callback
func WithNotify(fn NotifyFn) RetrierOption {
	return func(r *ExponentialRetrier) {
		r.notify = fn
	}
}
