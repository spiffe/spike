//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package env

import (
	"os"
	"strconv"
	"time"

	appEnv "github.com/spiffe/spike-sdk-go/config/env"
)

// DatabaseJournalMode returns the SQLite journal mode to use.
// It can be configured using the SPIKE_NEXUS_DB_JOURNAL_MODE environment
// variable.
//
// If the environment variable is not set, it defaults to "WAL"
// (Write-Ahead Logging).
func DatabaseJournalMode() string {
	s := os.Getenv(appEnv.NexusDBJournalMode)
	if s != "" {
		return s
	}
	return "WAL"
}

// DatabaseBusyTimeoutMs returns the SQLite busy timeout in milliseconds.
// It can be configured using the SPIKE_NEXUS_DB_BUSY_TIMEOUT_MS environment
// variable. The value must be a positive integer.
//
// If the environment variable is not set or contains an invalid value,
// it defaults to 5000 milliseconds (5 seconds).
func DatabaseBusyTimeoutMs() int {
	p := os.Getenv(appEnv.NexusDBBusyTimeoutMS)
	if p != "" {
		bt, err := strconv.Atoi(p)
		if err == nil && bt > 0 {
			return bt
		}
	}

	return 5000
}

// DatabaseMaxOpenConns returns the maximum number of open database connections.
// It can be configured using the SPIKE_NEXUS_DB_MAX_OPEN_CONNS environment
// variable. The value must be a positive integer.
//
// If the environment variable is not set or contains an invalid value,
// it defaults to 10 connections.
func DatabaseMaxOpenConns() int {
	p := os.Getenv(appEnv.NexusDBMaxOpenConns)
	if p != "" {
		moc, err := strconv.Atoi(p)
		if err == nil && moc > 0 {
			return moc
		}
	}

	return 10
}

// DatabaseMaxIdleConns returns the maximum number of idle database connections.
// It can be configured using the SPIKE_NEXUS_DB_MAX_IDLE_CONNS environment
// variable. The value must be a positive integer.
//
// If the environment variable is not set or contains an invalid value,
// it defaults to 5 connections.
func DatabaseMaxIdleConns() int {
	p := os.Getenv(appEnv.NexusDBMaxIdleConns)
	if p != "" {
		mic, err := strconv.Atoi(p)
		if err == nil && mic > 0 {
			return mic
		}
	}

	return 5
}

// DatabaseConnMaxLifetimeSec returns the maximum lifetime duration for a
// database connection. It can be configured using the
// SPIKE_NEXUS_DB_CONN_MAX_LIFETIME environment variable.
// The value should be a valid Go duration string (e.g., "1h", "30m").
//
// If the environment variable is not set or contains an invalid duration,
// it defaults to 1 hour.
func DatabaseConnMaxLifetimeSec() time.Duration {
	p := os.Getenv(appEnv.NexusDBConnMaxLifetime)
	if p != "" {
		d, err := time.ParseDuration(p)
		if err == nil {
			return d
		}
	}

	return time.Hour
}

// DatabaseOperationTimeout returns the duration to use for database operations.
// It can be configured using the SPIKE_NEXUS_DB_OPERATION_TIMEOUT environment
// variable. The value should be a valid Go duration string (e.g., "10s", "1m").
//
// If the environment variable is not set or contains an invalid duration,
// it defaults to 15 seconds.
func DatabaseOperationTimeout() time.Duration {
	p := os.Getenv(appEnv.NexusDBOperationTimeout)
	if p != "" {
		d, err := time.ParseDuration(p)
		if err == nil {
			return d
		}
	}

	return 15 * time.Second
}

// DatabaseInitializationTimeout returns the duration to wait for database
// initialization.
//
// The timeout is read from the environment variable
// `SPIKE_NEXUS_DB_INITIALIZATION_TIMEOUT`. If this variable is set and its
// value can be parsed as a duration (e.g., "1m30s"), it is used.
// Otherwise, the function defaults to a timeout of 30 seconds.
func DatabaseInitializationTimeout() time.Duration {
	p := os.Getenv(appEnv.NexusDBInitializationTimeout)
	if p != "" {
		d, err := time.ParseDuration(p)
		if err == nil {
			return d
		}
	}
	return 30 * time.Second
}

// DatabaseSkipSchemaCreation determines if schema creation should be skipped.
// It checks the "SPIKE_NEXUS_DB_SKIP_SCHEMA_CREATION" env variable to decide.
// If the env variable is set and its value is "true", it returns true.
// Otherwise, it returns false.
func DatabaseSkipSchemaCreation() bool {
	p := os.Getenv(appEnv.NexusDBSkipSchemaCreation)
	if p != "" {
		s, err := strconv.ParseBool(p)
		if err == nil {
			return s
		}
	}
	return false
}
