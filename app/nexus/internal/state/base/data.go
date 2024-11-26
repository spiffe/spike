//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package base

import (
	"sync"

	"github.com/spiffe/spike/app/nexus/internal/env"
	"github.com/spiffe/spike/app/nexus/internal/state/entity/data"
	"github.com/spiffe/spike/pkg/store"
)

var (
	rootKey   string
	rootKeyMu sync.RWMutex
)

var (
	adminSigningToken   string
	adminSigningTokenMu sync.RWMutex
)

var (
	kv = store.NewKV(store.KVConfig{
		MaxSecretVersions: env.MaxSecretVersions(),
	})
	kvMu sync.RWMutex
)

var policies sync.Map

var (
	adminRecoveryMetadata   data.RecoveryMetadata
	adminRecoveryMetadataMu sync.RWMutex
)
