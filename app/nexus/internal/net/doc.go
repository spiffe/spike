//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

// Package net provides network server utilities for SPIKE Nexus.
//
// This package includes functions for initializing and starting the
// TLS-secured HTTP server that handles incoming requests for secret
// management, policy administration, and operator operations. The server is
// configured with mTLS authentication.
package net
