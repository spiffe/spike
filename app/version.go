//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package app

import _ "embed"

// Version contains the application version string loaded from VERSION.txt at
// compile time. This value is embedded into the binary and used for version
// reporting in CLI output and logs.
//
//go:embed VERSION.txt
var Version string
