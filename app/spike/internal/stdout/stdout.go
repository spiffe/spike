//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

// Package stdout provides utilities for printing formatted messages to
// standard output. It contains functions for displaying notification and
// status messages to users.
package stdout

import "fmt"

// PrintNotReady prints a message indicating that SPIKE is not initialized
// and provides instructions for troubleshooting and recovery.
// The message includes suggestions to wait, check logs, and information about
// manual bootstrapping if the initialization problem persists.
func PrintNotReady() {
	fmt.Println("!")
	fmt.Println("!  SPIKE is not initialized.")
	fmt.Println("!  Wait a few seconds and try again.")
	fmt.Println("!  Also, check out SPIKE Nexus logs.")
	fmt.Println("!")
	fmt.Println("!  If the problem persists, you may need to")
	fmt.Println("!  manually bootstrap via `spike operator restore`.")
	fmt.Println("!")
	fmt.Println("!  Please check out https://spike.ist/ for additional")
	fmt.Println("!  recovery and restoration information.")
	fmt.Println("!")
}
