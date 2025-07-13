//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package log

import (
	"log"
	"log/slog"
	"os"
	"strings"
	"sync"
)

var logger *slog.Logger
var loggerMutex sync.Mutex

// Log returns a thread-safe singleton instance of slog.Logger configured for
// JSON output. If the logger hasn't been initialized, it creates a new instance
// with the log level specified by the environment. Subsequent calls return the
// same logger instance.
func Log() *slog.Logger {
	loggerMutex.Lock()
	defer loggerMutex.Unlock()

	if logger != nil {
		return logger
	}

	opts := &slog.HandlerOptions{
		Level: Level(),
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

// FatalLn logs a message at Fatal level with a line feed and then calls
// os.Exit(1).
func FatalLn(args ...any) {
	log.Fatalln(args...)
}

// Level returns the logging level for the SPIKE components.
//
// It reads from the SPIKE_SYSTEM_LOG_LEVEL environment variable and
// converts it to the corresponding slog.Level value.
// Valid values (case-insensitive) are:
//   - "DEBUG": returns slog.LevelDebug
//   - "INFO": returns slog.LevelInfo
//   - "WARN": returns slog.LevelWarn
//   - "ERROR": returns slog.LevelError
//
// If the environment variable is not set or contains an invalid value,
// it returns the default level slog.LevelWarn.
func Level() slog.Level {
	level := os.Getenv("SPIKE_SYSTEM_LOG_LEVEL")
	level = strings.ToUpper(level)

	switch level {
	case "DEBUG":
		return slog.LevelDebug // -4
	case "INFO":
		return slog.LevelInfo // 0
	case "WARN":
		return slog.LevelWarn // 4
	case "ERROR":
		return slog.LevelError // 8
	default:
		return slog.LevelWarn // 4
	}
}
