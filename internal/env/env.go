//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

// Package env provides utilities for managing environment variables
// and configurations specific to the SPIKE system.
//
// This package includes functionality to retrieve and interpret
// environment variable settings such as logging levels for the
// SPIKE components.
package env

import (
	"log/slog"
	"os"
	"strings"
)

// LogLevel returns the logging level for the SPIKE components.
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
func LogLevel() slog.Level {
	level := os.Getenv("SPIKE_SYSTEM_LOG_LEVEL")
	level = strings.ToUpper(level)

	switch level {
	case "DEBUG":
		return slog.LevelDebug
	case "INFO":
		return slog.LevelInfo
	case "WARN":
		return slog.LevelWarn
	case "ERROR":
		return slog.LevelError
	default:
		return slog.LevelWarn
	}
}
