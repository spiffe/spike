//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package system

import (
	"log"
	"os"
	"os/signal"
	"syscall"
)

// KeepAlive blocks the current goroutine until it receives either a
// SIGINT (Ctrl+C) or SIGTERM signal, enabling graceful shutdown of the
// application. Upon receiving a termination signal, it logs the signal type
// and begins the shutdown process.
//
// The function creates a buffered channel to handle OS signals and uses
// signal.Notify to register for SIGINT and SIGTERM signals. It then blocks
// until a signal is received.
//
// This is typically used in the main function to prevent the program from
// exiting immediately and to ensure proper cleanup when the program is
// terminated.
//
// Example usage:
//
//	func main() {
//	    // Initialize your application
//	    setupApp()
//
//	    // Keep the application running until shutdown signal
//	    KeepAlive()
//
//	    // Perform cleanup
//	    cleanup()
//	}
func KeepAlive() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigChan

	log.Printf("\nReceived %v signal, shutting down gracefully...\n", sig)
}
