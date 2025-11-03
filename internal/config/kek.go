//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package config

import (
	"os"
	"strconv"
)

// KEK rotation configuration environment variables and defaults

const (
	// EnvKEKRotationEnabled enables KEK rotation feature
	EnvKEKRotationEnabled = "SPIKE_KEK_ROTATION_ENABLED"

	// EnvKEKRotationDays is the number of days before KEK rotation
	EnvKEKRotationDays = "SPIKE_KEK_ROTATION_DAYS"

	// EnvKEKMaxWraps is the maximum number of wraps before KEK rotation
	EnvKEKMaxWraps = "SPIKE_KEK_MAX_WRAPS"

	// EnvKEKGraceDays is the grace period in days for old KEKs
	EnvKEKGraceDays = "SPIKE_KEK_GRACE_DAYS"

	// EnvKEKLazyRewrapEnabled enables lazy rewrapping on read
	EnvKEKLazyRewrapEnabled = "SPIKE_KEK_LAZY_REWRAP_ENABLED"

	// EnvKEKMaxRewrapQPS is the max rewrap rate in QPS
	EnvKEKMaxRewrapQPS = "SPIKE_KEK_MAX_REWRAP_QPS"

	// Default values
	DefaultKEKRotationEnabled   = false // Disabled by default for gradual rollout
	DefaultKEKRotationDays      = 90
	DefaultKEKMaxWraps          = 20_000_000
	DefaultKEKGraceDays         = 180
	DefaultKEKLazyRewrapEnabled = true
	DefaultKEKMaxRewrapQPS      = 100
)

// KEKRotationEnabled returns whether KEK rotation is enabled
func KEKRotationEnabled() bool {
	val := os.Getenv(EnvKEKRotationEnabled)
	if val == "" {
		return DefaultKEKRotationEnabled
	}
	enabled, err := strconv.ParseBool(val)
	if err != nil {
		return DefaultKEKRotationEnabled
	}
	return enabled
}

// KEKRotationDays returns the KEK rotation period in days
func KEKRotationDays() int {
	val := os.Getenv(EnvKEKRotationDays)
	if val == "" {
		return DefaultKEKRotationDays
	}
	days, err := strconv.Atoi(val)
	if err != nil || days <= 0 {
		return DefaultKEKRotationDays
	}
	return days
}

// KEKMaxWraps returns the maximum number of wraps before rotation
func KEKMaxWraps() int64 {
	val := os.Getenv(EnvKEKMaxWraps)
	if val == "" {
		return DefaultKEKMaxWraps
	}
	wraps, err := strconv.ParseInt(val, 10, 64)
	if err != nil || wraps <= 0 {
		return DefaultKEKMaxWraps
	}
	return wraps
}

// KEKGraceDays returns the grace period for old KEKs in days
func KEKGraceDays() int {
	val := os.Getenv(EnvKEKGraceDays)
	if val == "" {
		return DefaultKEKGraceDays
	}
	days, err := strconv.Atoi(val)
	if err != nil || days <= 0 {
		return DefaultKEKGraceDays
	}
	return days
}

// KEKLazyRewrapEnabled returns whether lazy rewrapping is enabled
func KEKLazyRewrapEnabled() bool {
	val := os.Getenv(EnvKEKLazyRewrapEnabled)
	if val == "" {
		return DefaultKEKLazyRewrapEnabled
	}
	enabled, err := strconv.ParseBool(val)
	if err != nil {
		return DefaultKEKLazyRewrapEnabled
	}
	return enabled
}

// KEKMaxRewrapQPS returns the maximum rewrap rate in queries per second
func KEKMaxRewrapQPS() int {
	val := os.Getenv(EnvKEKMaxRewrapQPS)
	if val == "" {
		return DefaultKEKMaxRewrapQPS
	}
	qps, err := strconv.Atoi(val)
	if err != nil || qps <= 0 {
		return DefaultKEKMaxRewrapQPS
	}
	return qps
}
