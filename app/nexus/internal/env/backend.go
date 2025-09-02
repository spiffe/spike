//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package env

import (
	"os"
	"strings"
)

// StoreType represents the type of backend storage to use.
type StoreType string

const (
	// Lite mode
	// This mode converts SPIKE to an encryption-as-a-service app.
	// It is used to store secrets in S3-compatible mediums (such as Minio)
	// without actually persisting them to a backing store.
	// In this mode SPIKE policies are "minimally" enforced, and the recommended
	// way to manage RBAC is to use the object storage's policy rules instead.
	Lite StoreType = "lite"

	// Sqlite indicates a SQLite database storage backend
	// This is the default backing store. SPIKE_NEXUS_BACKEND_STORE environment
	// variable can override it.
	Sqlite StoreType = "sqlite"

	// Memory indicates an in-memory storage backend
	// This mode is not recommended for production use as SPIKE will NOT rely on
	// SPIKE Keeper instances for Disaster Recovery and Redundancy.
	Memory StoreType = "memory"
)

// BackendStoreType determines which storage backend type to use based on the
// SPIKE_NEXUS_BACKEND_STORE environment variable. The value is
// case-insensitive.
//
// Valid values are:
//   - "lite": Lite mode that does not use any backing store
//   - "sqlite": Uses SQLite database storage
//   - "memory": Uses in-memory storage
//
// If the environment variable is not set or contains an invalid value,
// it defaults to SQLite.
func BackendStoreType() StoreType {
	st := os.Getenv("SPIKE_NEXUS_BACKEND_STORE")

	switch strings.ToLower(st) {
	case string(Lite):
		return Lite
	case string(Sqlite):
		return Sqlite
	case string(Memory):
		return Memory
	default:
		return Sqlite
	}
}
