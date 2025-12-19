//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

// Package net provides network server utilities for SPIKE Keeper.
//
// This package includes functions for initializing and starting the
// TLS-secured HTTP server that handles incoming requests from SPIKE Nexus and
// SPIKE Bootstrap. The server is configured with mTLS authentication and
// restricted to accepting connections only from authorized peers.
package net
