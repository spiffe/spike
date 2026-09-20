//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

// Package app holds the values shared by every SPIKE binary.
//
// Today that is the application version string, which is embedded from
// VERSION.txt at build time and reported by Nexus, Pilot, Keeper, and
// Bootstrap alike. Component-specific configuration does not belong here; it
// lives under each component's own internal packages.
package app
