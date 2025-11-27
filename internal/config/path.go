//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package config

import "sync"

// Cached directory paths and sync.Once for one-time initialization.
var (
	nexusDataPath     string
	nexusDataOnce     sync.Once
	pilotRecoveryPath string
	pilotRecoveryOnce sync.Once
)
