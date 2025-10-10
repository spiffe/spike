//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

// Package net provides network communication utilities for SPIKE Bootstrap.
//
// This package includes functions for creating mTLS clients, preparing
// payloads, and posting requests to SPIKE Keeper and SPIKE Nexus instances.
// It handles both shard contribution requests to keepers and verification
// requests to Nexus during the bootstrap process.
package net
