//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

// Package format defines the output formats that the SPIKE Pilot commands
// support (human-readable, JSON, and YAML) and the shared --format flag
// that selects between them.
//
// Commands that render structured output register the flag with
// AddFormatFlag and resolve the operator's choice with GetFormat, so that
// every command accepts the same format names and aliases.
package format
