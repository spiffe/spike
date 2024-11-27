//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package retry

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/stretchr/testify/assert"
)

func TestExponentialRetrier_Success(t *testing.T) {
    retrier := NewExponentialRetrier()
    
    // Operation that succeeds immediately
    err := retrier.RetryWithBackoff(context.Background(), func() error {
        return nil
    })
    
    assert.NoError(t, err)
}

func TestExponentialRetrier_EventualSuccess(t *testing.T) {
    attempts := 0
    maxAttempts := 3
    
    retrier := NewExponentialRetrier(
        WithBackOffOptions(
            WithInitialInterval(1 * time.Millisecond),
            WithMaxInterval(5 * time.Millisecond),
        ),
    )
    
    err := retrier.RetryWithBackoff(context.Background(), func() error {
        attempts++
        if attempts < maxAttempts {
            return errors.New("temporary error")
        }
        return nil
    })
    
    assert.NoError(t, err)
    assert.Equal(t, maxAttempts, attempts)
}

func TestExponentialRetrier_Failure(t *testing.T) {
    retrier := NewExponentialRetrier(
        WithBackOffOptions(
            WithMaxElapsedTime(10 * time.Millisecond),
            WithInitialInterval(1 * time.Millisecond),
        ),
    )
    
    expectedErr := errors.New("persistent error")
    err := retrier.RetryWithBackoff(context.Background(), func() error {
        return expectedErr
    })
    
    assert.Error(t, err)
    assert.Equal(t, expectedErr, err)
}

func TestExponentialRetrier_ContextCancellation(t *testing.T) {
    ctx, cancel := context.WithCancel(context.Background())
    
    retrier := NewExponentialRetrier(
        WithBackOffOptions(
            WithInitialInterval(100 * time.Millisecond),
        ),
    )
    
    go func() {
        time.Sleep(50 * time.Millisecond)
        cancel()
    }()
    
    err := retrier.RetryWithBackoff(ctx, func() error {
        return errors.New("keep retrying")
    })
    
    assert.Error(t, err)
    assert.Equal(t, context.Canceled, err)
}

func TestExponentialRetrier_Notification(t *testing.T) {
    var notifications []time.Duration
    totalDurations := make([]time.Duration, 0)
    
    retrier := NewExponentialRetrier(
        WithNotify(func(_ error, duration, totalDuration time.Duration) {
            notifications = append(notifications, duration)
            totalDurations = append(totalDurations, totalDuration)
        }),
        WithBackOffOptions(
            WithInitialInterval(1 * time.Millisecond),
            WithMaxInterval(5 * time.Millisecond),
            WithMaxElapsedTime(20 * time.Millisecond),
        ),
    )
    
    attempts := 0
    _ = retrier.RetryWithBackoff(context.Background(), func() error {
        attempts++
        if attempts < 3 {
            return errors.New("temporary error")
        }
        return nil
    })
    
    assert.Equal(t, 2, len(notifications))
    assert.Equal(t, 2, len(totalDurations))
    
    // Verify that durations are increasing
    assert.Less(t, notifications[0], notifications[1])
    assert.Less(t, totalDurations[0], totalDurations[1])
}

func TestTypedRetrier_Success(t *testing.T) {
    baseRetrier := NewExponentialRetrier()
    typedRetrier := NewTypedRetrier[string](baseRetrier)
    
    expected := "success"
    result, err := typedRetrier.RetryWithBackoff(context.Background(), func() (string, error) {
        return expected, nil
    })
    
    assert.NoError(t, err)
    assert.Equal(t, expected, result)
}

func TestTypedRetrier_Failure(t *testing.T) {
    baseRetrier := NewExponentialRetrier(
        WithBackOffOptions(
            WithMaxElapsedTime(10 * time.Millisecond),
            WithInitialInterval(1 * time.Millisecond),
        ),
    )
    typedRetrier := NewTypedRetrier[int](baseRetrier)
    
    expectedErr := errors.New("typed error")
    result, err := typedRetrier.RetryWithBackoff(context.Background(), func() (int, error) {
        return 0, expectedErr
    })
    
    assert.Error(t, err)
    assert.Equal(t, expectedErr, err)
    assert.Equal(t, 0, result)
}

func TestBackOffOptions(t *testing.T) {
    initialInterval := 100 * time.Millisecond
    maxInterval := 1 * time.Second
    maxElapsedTime := 5 * time.Second
    multiplier := 2.5
    
    retrier := NewExponentialRetrier(
        WithBackOffOptions(
            WithInitialInterval(initialInterval),
            WithMaxInterval(maxInterval),
            WithMaxElapsedTime(maxElapsedTime),
            WithMultiplier(multiplier),
        ),
    )
    
    // Access the backoff configuration
    b := retrier.newBackOff().(*backoff.ExponentialBackOff)
    
    assert.Equal(t, initialInterval, b.InitialInterval)
    assert.Equal(t, maxInterval, b.MaxInterval)
    assert.Equal(t, maxElapsedTime, b.MaxElapsedTime)
    assert.Equal(t, multiplier, b.Multiplier)
}

func TestDefaultSettings(t *testing.T) {
	retrier := NewExponentialRetrier()
	b := retrier.newBackOff().(*backoff.ExponentialBackOff)

	assert.Equal(t, defaultInitialInterval, b.InitialInterval)
	assert.Equal(t, defaultMaxInterval, b.MaxInterval)
	assert.Equal(t, defaultMaxElapsedTime, b.MaxElapsedTime)
	assert.Equal(t, defaultMultiplier, b.Multiplier)
}