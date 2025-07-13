//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package persist

import (
	"github.com/spiffe/spike/app/nexus/internal/env"
	"time"
)

// Options defines SQLite-specific configuration options
type Options struct {
	// DataDir specifies the directory where the SQLite database file
	// will be stored
	DataDir string

	// DatabaseFile specifies the name of the SQLite database file
	DatabaseFile string

	// JournalMode specifies the SQLite journal mode
	// (DELETE, WAL, MEMORY, etc.)
	JournalMode string

	// BusyTimeoutMs specifies the busy timeout in milliseconds
	BusyTimeoutMs int

	// MaxOpenConns specifies the maximum number of open connections
	MaxOpenConns int

	// MaxIdleConns specifies the maximum number of idle connections
	MaxIdleConns int

	// ConnMaxLifetime specifies the maximum amount of time
	// a connection may be reused
	ConnMaxLifetime time.Duration
}

// DefaultOptions returns the default SQLite options
func DefaultOptions() *Options {
	return &Options{
		DataDir:         ".data",
		DatabaseFile:    "spike.db",
		JournalMode:     env.DatabaseJournalMode(),
		BusyTimeoutMs:   env.DatabaseBusyTimeoutMs(),
		MaxOpenConns:    env.DatabaseMaxOpenConns(),
		MaxIdleConns:    env.DatabaseMaxIdleConns(),
		ConnMaxLifetime: env.DatabaseConnMaxLifetimeSec(),
	}
}
