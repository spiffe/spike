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
	fmt.Fprintln(os.Stderr, "!\n")
	fmt.Fprintln(os.Stderr, "!  SPIKE is not initialized.\n")
	fmt.Fprintln(os.Stderr, "!  Wait a few seconds and try again.\n")
	fmt.Fprintln(os.Stderr, "!  Also, check out SPIKE Nexus logs.\n")
	fmt.Fprintln(os.Stderr, "!\n")
	fmt.Fprintln(os.Stderr, "!  If the problem persists, you may need to\n")
	fmt.Fprintln(os.Stderr, "!  manually bootstrap via `spike operator restore`.\n")
	fmt.Fprintln(os.Stderr, "!\n")
	fmt.Fprintln(os.Stderr, "!  Please check out https://spike.ist/ for additional\n")
	fmt.Fprintln(os.Stderr, "!  recovery and restoration information.\n")
	fmt.Fprintln(os.Stderr, "!\n")
}
