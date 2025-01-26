//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

// Package store provides a secure in-memory key-value store for managing secret
// data. The store supports versioning of secrets, allowing operations on
// specific versions and tracking deleted versions. It is designed for scenarios
// where secrets need to be securely managed, updated, and deleted.
package store
