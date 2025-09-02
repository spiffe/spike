//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

// Package recovery provides the functionality for securing and managing secrets
// using SPIKE Nexus. It includes methods for managing the bootstrapping and
// recovery process, such as interacting with keepers to contribute or recover
// shares and ensuring the initialization of the system.
//
// The package handles critical operations related to secret sharing,
// managing root keys, and maintaining system state. It ensures reliability
// during initialization and provides mechanisms to detect and handle failures
// in contributing or retrieving secret shares.
package recovery
