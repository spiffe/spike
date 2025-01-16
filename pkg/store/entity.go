//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package store

import "time"

// Version represents a single version of a secret's data along with its
// metadata. Each version maintains its own set of key-value pairs and tracking
// information.
type Version struct {
	// Data contains the actual key-value pairs stored in this version
	Data map[string]string

	// CreatedTime is when this version was created
	CreatedTime time.Time

	// Version is the numeric identifier for this version
	Version int

	// DeletedTime indicates when this version was marked as deleted
	// A nil value means the version is active/not deleted
	DeletedTime *time.Time
}

// Metadata tracks control information for a secret and its versions.
// It maintains version boundaries and timestamps for the overall secret.
type Metadata struct {
	// CurrentVersion is the newest/latest version number of the secret
	CurrentVersion int

	// OldestVersion is the oldest available version number of the secret
	OldestVersion int

	// CreatedTime is when the secret was first created
	CreatedTime time.Time

	// UpdatedTime is when the secret was last modified
	UpdatedTime time.Time

	// MaxVersions is the maximum number of versions to retain
	// When exceeded, older versions are automatically pruned
	MaxVersions int
}

// Secret represents a versioned collection of key-value pairs stored at a
// specific path. It maintains both the version history and metadata about the
// collection as a whole.
type Secret struct {
	// Versions maps version numbers to their corresponding Version objects
	Versions map[int]Version

	// Metadata contains control information about this secret
	Metadata Metadata
}

// TODO: we don't need keyrecoverydata anymore.

// KeyRecoveryData represents the complete set of recovery information
// for the root key and its Shamir shards.
type KeyRecoveryData struct {
	// RootKey contains the root encryption key
	RootKey []byte

	// MinShards is the minimum number of shards needed to reconstruct the
	// root key
	MinShards int

	// TotalShards is the total number of shards that were created
	TotalShards int

	// Shards contains all the Shamir shards of the root key.
	// The slice index (0-based) corresponds to the shard index
	Shards [][]byte

	// CreatedTime is when this recovery data was first created
	CreatedTime time.Time

	// UpdatedTime is when this recovery data was last modified
	UpdatedTime time.Time
}
