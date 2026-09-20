//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\ Copyright 2024-present SPIKE contributors.
// \\\\\ SPDX-License-Identifier: Apache-2.0

// Package base provides the core routing logic for the SPIKE application's
// HTTP server. It dynamically resolves incoming HTTP requests to the
// appropriate handlers based on their URL paths and methods. This package
// ensures flexibility and extensibility in supporting various API actions and
// paths within SPIKE's ecosystem.
package base
