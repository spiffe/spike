//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package persist

import (
	"fmt"
	"time"

	sdkErrors "github.com/spiffe/spike-sdk-go/errors"
	"github.com/spiffe/spike/app/nexus/internal/state/backend"
)

// ParseOptions parses and validates the provided backend options
func ParseOptions(opts map[backend.DatabaseConfigKey]any) (*Options, *sdkErrors.SDKError) {
	if opts == nil {
		return DefaultOptions(), nil
	}

	sqliteOpts := &Options{}

	// Parse each field from the map
	if dataDir, ok := opts[backend.KeyDataDir].(string); ok {
		sqliteOpts.DataDir = dataDir
	}
	if dbFile, ok := opts[backend.KeyDatabaseFile].(string); ok {
		sqliteOpts.DatabaseFile = dbFile
	}
	if journalMode, ok := opts[backend.KeyJournalMode].(string); ok {
		sqliteOpts.JournalMode = journalMode
	}
	if busyTimeout, ok := opts[backend.KeyBusyTimeoutMs].(int); ok {
		sqliteOpts.BusyTimeoutMs = busyTimeout
	}
	if maxOpen, ok := opts[backend.KeyMaxOpenConns].(int); ok {
		sqliteOpts.MaxOpenConns = maxOpen
	}
	if maxIdle, ok := opts[backend.KeyMaxIdleConns].(int); ok {
		sqliteOpts.MaxIdleConns = maxIdle
	}
	if lifetime, ok := opts[backend.KeyConnMaxLifetimeSeconds].(time.Duration); ok {
		sqliteOpts.ConnMaxLifetime = lifetime
	}

	// Apply defaults for zero values
	if sqliteOpts.DataDir == "" {
		sqliteOpts.DataDir = DefaultOptions().DataDir
	}
	if sqliteOpts.DatabaseFile == "" {
		sqliteOpts.DatabaseFile = DefaultOptions().DatabaseFile
	}
	if sqliteOpts.JournalMode == "" {
		sqliteOpts.JournalMode = DefaultOptions().JournalMode
	}
	if sqliteOpts.BusyTimeoutMs == 0 {
		sqliteOpts.BusyTimeoutMs = DefaultOptions().BusyTimeoutMs
	}
	if sqliteOpts.MaxOpenConns == 0 {
		sqliteOpts.MaxOpenConns = DefaultOptions().MaxOpenConns
	}
	if sqliteOpts.MaxIdleConns == 0 {
		sqliteOpts.MaxIdleConns = DefaultOptions().MaxIdleConns
	}
	if sqliteOpts.ConnMaxLifetime == 0 {
		sqliteOpts.ConnMaxLifetime = DefaultOptions().ConnMaxLifetime
	}

	// Validate options
	if sqliteOpts.MaxIdleConns > sqliteOpts.MaxOpenConns {
		return nil,
			fmt.Errorf(
				"MaxIdleConns (%d) cannot be greater than MaxOpenConns (%d)",
				sqliteOpts.MaxIdleConns, sqliteOpts.MaxOpenConns,
			)
	}

	return sqliteOpts, nil
}
