//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

// Package operator provides HTTP route handlers for SPIKE Nexus operator
// operations.
//
// This package implements recovery and restore endpoints that allow operators
// to recover Nexus from disaster scenarios. The recover endpoint retrieves
// recovery shards for SPIKE Pilot, and the restore endpoint accepts recovery
// shards to reconstruct the root key and restore Nexus operation.
package operator
