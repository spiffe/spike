//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package log

import (
	"encoding/json"
	"log"
	"log/slog"
	"os"
	"sync"
	"time"

	"github.com/spiffe/spike/internal/env"
)

var logger *slog.Logger
var loggerMutex sync.Mutex

// Log returns a thread-safe singleton instance of slog.Logger configured for JSON output.
// If the logger hasn't been initialized, it creates a new instance with the log level
// specified by the environment. Subsequent calls return the same logger instance.
func Log() *slog.Logger {
	loggerMutex.Lock()
	defer loggerMutex.Unlock()

	if logger != nil {
		return logger
	}

	opts := &slog.HandlerOptions{
		Level: env.LogLevel(),
	}

	handler := slog.NewJSONHandler(os.Stdout, opts)

	logger = slog.New(handler)
	return logger
}

// Fatal logs a message at Fatal level and then calls os.Exit(1).
func Fatal(msg string) {
	log.Fatal(msg)
}

// FatalF logs a formatted message at Fatal level and then calls os.Exit(1).
// It follows the printf formatting rules.
func FatalF(format string, args ...any) {
	log.Fatalf(format, args...)
}

// FatalLn logs a message at Fatal level with a line feed and then calls os.Exit(1).
func FatalLn(args ...any) {
	log.Fatalln(args...)
}

// AuditEntry represents a single audit log entry containing information about
// user actions within the system.
type AuditEntry struct {
	// Timestamp indicates when the audited action occurred
	Timestamp time.Time

	// UserId identifies the user who performed the action
	UserId string

	// Action describes what operation was performed
	Action string

	// Resource identifies the object or entity that was acted upon
	Resource string

	// SessionID links the action to a specific user session
	SessionID string
}

// Audit logs an audit entry as JSON to the standard log output.
// If JSON marshaling fails, it logs an error using the structured logger
// but continues execution.
func Audit(entry AuditEntry) {
	body, err := json.Marshal(entry)
	if err != nil {
		Log().Error("Audit",
			"msg", "Problem marshalling audit entry",
			"err", err.Error())
		return
	}

	log.Println(string(body))
}
