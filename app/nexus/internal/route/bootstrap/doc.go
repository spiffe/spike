//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

// Package bootstrap provides HTTP route handlers for SPIKE Bootstrap
// verification endpoints.
//
// This package implements the verification endpoint that allows SPIKE
// Bootstrap to verify that SPIKE Nexus has been properly initialized with the
// root key by sending encrypted data and validating the decryption result.
package bootstrap
