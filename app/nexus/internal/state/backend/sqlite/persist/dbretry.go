//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package persist

import (
	"context"
	"strings"
	"time"

	"github.com/spiffe/spike-sdk-go/config/env"
)

// Retry pacing for transient SQLite failures. The overall duration is
// bounded by the operation deadline (see operationContext), so these
// values only shape how densely attempts are packed inside that window.
const (
	dbRetryInitialInterval = 100 * time.Millisecond
	dbRetryMaxInterval     = 2 * time.Second
)

// operationContext bounds a database operation with the configured
// SPIKE_NEXUS_DB_OPERATION_TIMEOUT deadline. A zero or negative timeout
// disables the deadline, in which case the parent context is returned
// unchanged with a no-op cancel function.
//
// Parameters:
//   - ctx: The parent context.
//
// Returns:
//   - context.Context: The possibly deadline-bound context.
//   - context.CancelFunc: The cancel function; callers must defer it.
func operationContext(
	ctx context.Context,
) (context.Context, context.CancelFunc) {
	timeout := env.DatabaseOperationTimeoutVal()
	if timeout <= 0 {
		return ctx, func() {}
	}
	return context.WithTimeout(ctx, timeout)
}

// transientDBError reports whether err is a transient SQLite failure
// (SQLITE_BUSY or SQLITE_LOCKED) that is worth retrying. These surface
// when a competing transaction holds the database lock for longer than
// the driver's busy timeout. Anything else is treated as permanent.
//
// The check matches the driver's error strings rather than its typed
// error codes: the typed API (sqlite3.Error) only compiles with cgo,
// and the repository's lint pass runs with CGO_ENABLED=0. Error
// messages survive SDKError wrapping, so the whole chain is covered.
//
// Parameters:
//   - err: The error to inspect. May be nil.
//
// Returns:
//   - bool: true if the error is a transient SQLite lock error.
func transientDBError(err error) bool {
	if err == nil {
		return false
	}

	// SQLITE_BUSY renders as "database is locked" and SQLITE_LOCKED as
	// "database table is locked" in the mattn/go-sqlite3 driver.
	msg := err.Error()
	return strings.Contains(msg, "database is locked") ||
		strings.Contains(msg, "database table is locked")
}
