//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package base

import (
	"github.com/spiffe/spike/app/nexus/internal/state/entity/data"
	"github.com/spiffe/spike/app/nexus/internal/state/store"
	"sync"
)

var (
	rootKey   string
	rootKeyMu sync.RWMutex

	adminToken   string
	adminTokenMu sync.RWMutex

	kv   = store.NewKV()
	kvMu sync.RWMutex

	adminCredentials   data.Credentials
	adminCredentialsMu sync.RWMutex
)
