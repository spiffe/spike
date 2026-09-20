//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"fmt"

	spike "github.com/spiffe/spike-sdk-go/api"
	"github.com/spiffe/spike-sdk-go/log"
)

const appName = "SPIKE Demo"

// say writes a line to stdout and terminates the program if the write fails.
//
// The demo has no purpose if it cannot show its output, so a failed write is
// treated as fatal rather than ignored.
//
// Parameters:
//   - args: The values to print, formatted as fmt.Println would.
func say(args ...any) {
	if _, err := fmt.Println(args...); err != nil {
		log.FatalLn(appName, "message", "failed to write to stdout",
			"err", err.Error())
	}
}

func main() {
	say("SPIKE Demo")

	// Make sure you register the demo app SPIRE Server registration entry
	// first:
	// ./examples/consume-secrets/demo-register-entry.sh

	// https://pkg.go.dev/github.com/spiffe/spike-sdk-go/api#New
	api, connErr := spike.New() // Use the default Workload API Socket
	if connErr != nil {
		say("Error connecting to SPIKE Nexus:", connErr.Error())
		return
	}

	say("Connected to SPIKE Nexus.")

	// https://pkg.go.dev/github.com/spiffe/spike-sdk-go/api#Close
	defer func() {
		// Close the connection when done
		closeErr := api.Close()
		if closeErr != nil {
			say("Error closing connection:", closeErr.Error())
		}
	}()

	// The path to store/retrieve/update the secret.
	path := "tenants/demo/db/creds"

	ctx := context.Background()

	// Create a Secret
	// https://pkg.go.dev/github.com/spiffe/spike-sdk-go/api#PutSecret
	putErr := api.PutSecret(ctx, path, map[string]string{
		"username": "SPIKE",
		"password": "SPIKE_Rocks",
	})
	if putErr != nil {
		say("Error writing secret:", putErr.Error())
		return
	}

	// Read the Secret
	// https://pkg.go.dev/github.com/spiffe/spike-sdk-go/api#GetSecret
	secret, getErr := api.GetSecret(ctx, path)
	if getErr != nil {
		say("Error reading secret:", getErr.Error())
		return
	}

	if secret == nil {
		say("Secret not found.")
		return
	}

	say("Secret found:")
	for k, v := range secret.Data {
		say(k + ": " + v)
	}
}
