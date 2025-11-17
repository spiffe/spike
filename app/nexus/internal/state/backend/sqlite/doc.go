//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

// Package sqlite provides an encrypted SQLite-based implementation of a data
// store backend. It supports storing and loading encrypted secrets and admin
// tokens with versioning support.
package sqlite

import (
	// Imported for side effects.
	_ "github.com/mattn/go-sqlite3"
)
