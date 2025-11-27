//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

// Package store provides HTTP route handlers for SPIKE Keeper's shard
// management operations. It handles receiving shard contributions from SPIKE
// Bootstrap during initialization and serving shards to SPIKE Nexus during
// recovery. All operations enforce SPIFFE ID-based authorization and include
// audit logging.
package store
