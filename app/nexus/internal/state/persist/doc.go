//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

// Package persist provides functionality for managing and interacting with
// backend storage systems. It includes thread-safe operations to retrieve
// the initialized backend instance used by the application. This package
// supports multiple backend types, such as in-memory and SQLite, to cater
// to different storage requirements.
//
// The primary exported function, Backend, allows access to the current
// storage backend in a safe manner, ensuring data integrity during concurrent
// access. The backend type is determined based on the configured environment
// settings.
package persist
