//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

// Package env provides utility functions for environment variable management
// and configurations used within the Spike Keeper service.
//
// This package includes methods to retrieve values such as the TLS port
// required for the service. It ensures sensible default values when environment
// variables are not set, promoting reliability and ease of use.
package env
