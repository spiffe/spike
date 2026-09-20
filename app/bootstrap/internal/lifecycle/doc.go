//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

// Package lifecycle manages the bootstrap lifecycle for SPIKE.
//
// This package provides utilities for determining whether the bootstrap
// process should run and for marking bootstrap completion in Kubernetes
// environments using ConfigMaps. It coordinates multiple bootstrap
// instances so that bootstrap operations run exactly once per cluster.
package lifecycle
