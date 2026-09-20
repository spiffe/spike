//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

// Package config exposes the per-component version strings.
//
// All SPIKE components are built and released together, so NexusVersion,
// PilotVersion, KeeperVersion, and BootstrapVersion all resolve to the single
// application version embedded by the app package. Keeping one exported
// name per component lets a component report its version without knowing
// how the value is produced.
package config
