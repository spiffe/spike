//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

// Package secret provides functionality to securely manage secrets within the
// SPIKE ecosystem.
//
// This package includes commands for CRUD operations on secrets, leveraging
// SPIFFE identities for secure interactions with the SPIKE API. It abstracts
// the complexity of managing secret versions and ensuring trust, enabling
// developers to focus on their core application logic.
//
// The package is built using the cobra library to enable a command-line
// interface for secret management operations.
//
// Key Features:
// - Secure management of secrets and versions
// - SPIFFE-based authentication
// - Easy-to-use CLI commands for secret operations
package secret
