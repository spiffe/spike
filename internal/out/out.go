//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

// Package out provides utility functions for application initialization output,
// including banner display and memory locking operations. These functions are
// typically called during the startup phase of SPIKE applications to provide
// consistent initialization behavior across all components.
package out

import (
	"crypto/fips140"
	"fmt"

	"github.com/spiffe/spike-sdk-go/config/env"
	"github.com/spiffe/spike-sdk-go/log"
	"github.com/spiffe/spike-sdk-go/security/mem"
)

// PrintBanner outputs the application banner to standard output, including
// the application name, version, log level, and FIPS 140.3 status. The banner
// is only printed if the SPIKE_BANNER_ENABLED environment variable is set to
// true.
func PrintBanner(appName, appVersion string) {
	if !env.BannerEnabledVal() {
		return
	}

	fmt.Printf(`
   \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
 \\\\\ Copyright 2024-present SPIKE contributors.
\\\\\\\ SPDX-License-Identifier: Apache-2.0`+"\n\n"+
		"%s v%s. | LOG LEVEL: %s; FIPS 140.3 Enabled: %v\n\n",
		appName, appVersion, log.Level(), fips140.Enabled(),
	)
}

// LogMemLock attempts to lock the application's memory to prevent sensitive
// data from being swapped to disk. It logs the result of the operation. If
// memory locking succeeds, a success message is logged. If it fails, a warning
// is logged only if SPIKE_SHOW_MEMORY_WARNING is enabled.
func LogMemLock(appName string) {
	if mem.Lock() {
		log.Log().Info(
			appName,
			"message", "successfully locked memory",
		)
		return
	}
	if !env.ShowMemoryWarningVal() {
		return
	}
	log.Log().Info(
		appName,
		"message", "memory is not locked: please disable swap",
	)
}

// Preamble performs standard application initialization output by printing
// the application banner and attempting to lock memory. This function should
// be called during application startup.
func Preamble(appName, appVersion string) {
	PrintBanner(appName, appVersion)
	LogMemLock(appName)
}
