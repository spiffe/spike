//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package env

import (
	"os"
	"strconv"
	"strings"
	"time"
)

// TlsPort returns the TLS port for the Spike Nexus service.
// It reads from the SPIKE_NEXUS_TLS_PORT environment variable.
// If the environment variable is not set, it returns the default port ":8553".
func TlsPort() string {
	p := os.Getenv("SPIKE_NEXUS_TLS_PORT")
	if p != "" {
		return p
	}

	return ":8553"
}

// PollInterval returns the polling interval duration for the Spike Nexus service.
// It reads from the SPIKE_NEXUS_POLL_INTERVAL environment variable which should
// contain a valid duration string (e.g. "1h", "30m", "1h30m").
// If the environment variable is not set or contains an invalid duration,
// it returns the default interval of 5 minutes.
func PollInterval() time.Duration {
	p := os.Getenv("SPIKE_NEXUS_POLL_INTERVAL")
	if p != "" {
		d, err := time.ParseDuration(p)
		if err == nil {
			return d
		}
	}

	return 5 * time.Minute
}

// DatabaseOperationTimeout returns the duration to use for database operations.
// It can be configured using the SPIKE_NEXUS_DB_OPERATION_TIMEOUT environment variable.
// The value should be a valid Go duration string (e.g., "10s", "1m").
//
// If the environment variable is not set or contains an invalid duration,
// it defaults to 5 seconds.
func DatabaseOperationTimeout() time.Duration {
	p := os.Getenv("SPIKE_NEXUS_DB_OPERATION_TIMEOUT")
	if p != "" {
		d, err := time.ParseDuration(p)
		if err == nil {
			return d
		}
	}

	return 5 * time.Second
}

// MaxSecretVersions returns the maximum number of versions to retain for each secret.
// It reads from the SPIKE_NEXUS_MAX_SECRET_VERSIONS environment variable which should
// contain a positive integer value.
// If the environment variable is not set, contains an invalid integer, or specifies
// a non-positive value, it returns the default of 10 versions.
func MaxSecretVersions() int {
	p := os.Getenv("SPIKE_NEXUS_MAX_SECRET_VERSIONS")
	if p != "" {
		mv, err := strconv.Atoi(p)
		if err == nil && mv > 0 {
			return mv
		}
	}

	return 10
}

// StoreType represents the type of backend storage to use.
type StoreType string

const (
	// S3 indicates an Amazon S3 (or compatible) storage backend
	S3 StoreType = "s3"

	// Sqlite indicates a SQLite database storage backend
	Sqlite StoreType = "sqlite"

	// Memory indicates an in-memory storage backend
	Memory StoreType = "memory"
)

// BackendStoreType determines which storage backend type to use based on the
// SPIKE_NEXUS_BACKEND_STORE environment variable. The value is case-insensitive.
//
// Valid values are:
//   - "file": Uses file-based storage
//   - "sqlite": Uses SQLite database storage
//   - "memory": Uses in-memory storage
//
// If the environment variable is not set or contains an invalid value,
// it defaults to Memory.
func BackendStoreType() StoreType {
	st := os.Getenv("SPIKE_NEXUS_BACKEND_STORE")

	switch strings.ToLower(st) {
	case string(S3):
		return S3
	case string(Sqlite):
		return Sqlite
	case string(Memory):
		return Memory
	default:
		return Memory
	}
}

// DatabaseDir returns the directory path where database files should be stored.
// It can be configured using the SPIKE_NEXUS_DB_DATA_DIR environment variable.
//
// If the environment variable is not set, it defaults to "./.data".
func DatabaseDir() string {
	s := os.Getenv("SPIKE_NEXUS_DB_DATA_DIR")
	if s != "" {
		return s
	}
	return "./.data"
}

// DatabaseJournalMode returns the SQLite journal mode to use.
// It can be configured using the SPIKE_NEXUS_DB_JOURNAL_MODE environment variable.
//
// If the environment variable is not set, it defaults to "WAL" (Write-Ahead Logging).
func DatabaseJournalMode() string {
	s := os.Getenv("SPIKE_NEXUS_DB_JOURNAL_MODE")
	if s != "" {
		return s
	}
	return "WAL"
}

// DatabaseBusyTimeoutMs returns the SQLite busy timeout in milliseconds.
// It can be configured using the SPIKE_NEXUS_DB_BUSY_TIMEOUT_MS environment variable.
// The value must be a positive integer.
//
// If the environment variable is not set or contains an invalid value,
// it defaults to 5000 milliseconds (5 seconds).
func DatabaseBusyTimeoutMs() int {
	p := os.Getenv("SPIKE_NEXUS_DB_BUSY_TIMEOUT_MS")
	if p != "" {
		bt, err := strconv.Atoi(p)
		if err == nil && bt > 0 {
			return bt
		}
	}

	return 5000
}

// DatabaseMaxOpenConns returns the maximum number of open database connections.
// It can be configured using the SPIKE_NEXUS_DB_MAX_OPEN_CONNS environment variable.
// The value must be a positive integer.
//
// If the environment variable is not set or contains an invalid value,
// it defaults to 10 connections.
func DatabaseMaxOpenConns() int {
	p := os.Getenv("SPIKE_NEXUS_DB_MAX_OPEN_CONNS")
	if p != "" {
		moc, err := strconv.Atoi(p)
		if err == nil && moc > 0 {
			return moc
		}
	}

	return 10
}

// DatabaseMaxIdleConns returns the maximum number of idle database connections.
// It can be configured using the SPIKE_NEXUS_DB_MAX_IDLE_CONNS environment variable.
// The value must be a positive integer.
//
// If the environment variable is not set or contains an invalid value,
// it defaults to 5 connections.
func DatabaseMaxIdleConns() int {
	p := os.Getenv("SPIKE_NEXUS_DB_MAX_IDLE_CONNS")
	if p != "" {
		mic, err := strconv.Atoi(p)
		if err == nil && mic > 0 {
			return mic
		}
	}

	return 5
}

// DatabaseConnMaxLifetimeSec returns the maximum lifetime duration for a database connection.
// It can be configured using the SPIKE_NEXUS_DB_CONN_MAX_LIFETIME environment variable.
// The value should be a valid Go duration string (e.g., "1h", "30m").
//
// If the environment variable is not set or contains an invalid duration,
// it defaults to 1 hour.
func DatabaseConnMaxLifetimeSec() time.Duration {
	p := os.Getenv("SPIKE_NEXUS_DB_CONN_MAX_LIFETIME")
	if p != "" {
		d, err := time.ParseDuration(p)
		if err == nil {
			return d
		}
	}

	return time.Hour
}
