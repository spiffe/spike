//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

// Package env provides utilities for configuring and retrieving environmental
// settings for the SPIKE Nexus application. This includes determining the
// type of storage backend to use based on predefined environment variables.
// Valid storage backends include Amazon S3, SQLite, and in-memory storage.
//
// The package ensures a default storage type is used if no valid environment
// variable is set or if the specified type is not supported.
package env
