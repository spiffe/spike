//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package stdout

import (
	"fmt"
	"os"
	"sync"

	"github.com/spiffe/spike-sdk-go/log"
)

// notReadyCallCount tracks how many times PrintNotReady has been called.
// This enables progressive messaging: brief on the first call, detailed on
// the following calls.
var (
	notReadyCallCount int
	notReadyMu        sync.Mutex
)

// PrintNotReady prints a message indicating that SPIKE is not initialized.
//
// On the first call, it prints a brief message suggesting the user wait.
// On the following calls, it prints a more detailed message with
// troubleshooting steps and recovery instructions. This progressive approach
// avoids alarming users during normal startup delays while still providing
// help when there is a real problem.
func PrintNotReady() {
	const fName = "PrintNotReady"

	notReadyMu.Lock()
	notReadyCallCount++
	count := notReadyCallCount
	notReadyMu.Unlock()

	var msg string
	if count == 1 {
		msg = `
  SPIKE is not ready yet. Please wait a moment and try again.
`
	} else {
		msg = `
  SPIKE is not initialized.
  Wait a few seconds and try again.
  Also, check out SPIKE Nexus logs.

  If the problem persists, you may need to
  manually bootstrap via 'spike operator restore'.

  Please check out https://spike.ist/ for additional
  recovery and restoration information.
`
	}

	if _, err := fmt.Fprint(os.Stderr, msg); err != nil {
		// The Pilot cannot reach its own stderr; do not exit 0 as if the
		// user had been told.
		log.FatalLn(fName, "message", "failed to write to stderr",
			"err", err.Error())
	}
}
