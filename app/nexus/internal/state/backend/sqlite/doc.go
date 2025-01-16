//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package sqlite

import (
	_ "github.com/mattn/go-sqlite3"
)

// Package provides an encrypted SQLite-based implementation of a data store
// backend. It supports storing and loading encrypted secrets and admin tokens
// with versioning support.

// TODO: consider adding docs to other packages too.
