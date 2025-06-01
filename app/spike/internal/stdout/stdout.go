//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

// Package stdout provides utilities for printing formatted messages to
// standard output. It contains functions for displaying notification and
// status messages to users.
package stdout

import "fmt"
import "os"

// PrintNotReady prints a message indicating that SPIKE is not initialized
// and provides instructions for troubleshooting and recovery.
// The message includes suggestions to wait, check logs, and information about
// manual bootstrapping if the initialization problem persists.
func PrintNotReady() {
	_, err := fmt.Fprintf(os.Stderr, `!
	!	SPIKE is not initialized.
	!	Wait a few seconds and try again.
	!	Also, check out SPIKE Nexus logs.
	!
	!	If the problem persists, you may need to
	!	manually bootstrap via `+"`spike operator restore`"+`.
	!
	!	Please check out https://spike.ist/ for additional
	!	recovery and restoration information.
	!
	`)
	if err != nil {
		fmt.Println("Error writing to stderr:", err.Error())
		return
	}
}
