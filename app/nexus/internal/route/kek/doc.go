//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

// Package kek provides HTTP route handlers for KEK (Key Encryption Key)
// management operations in SPIKE Nexus. It handles KEK rotation, status
// queries, listing, and RMK (Root Master Key) rotation ceremonies.
//
// # Available Endpoints
//
// KEK Management:
//   - POST /v1/kek/rotate - Manually trigger KEK rotation
//   - GET /v1/kek/current - Get information about the current active KEK
//   - GET /v1/kek/list - List all KEKs with their status
//   - GET /v1/kek/stats - Get detailed rotation statistics
//
// RMK Management:
//   - POST /v1/rmk/rotate - Initiate RMK rotation ceremony (manual process)
//   - GET /v1/rmk/snapshot - Create a snapshot for RMK rotation
//
// # Security
//
// These endpoints require authentication and appropriate permissions.
// All operations are audited through the journal system.
package kek

