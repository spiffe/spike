//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"fmt"

	spike "github.com/spiffe/spike-sdk-go/api"
)

func main() {
	fmt.Println("SPIKE Demo")

	// Make sure you register the demo app SPIRE Server registration entry
	// first:
	// ./examples/consume-secrets/demo-register-entry.sh

	// https://pkg.go.dev/github.com/spiffe/spike-sdk-go/api#New
	api, connErr := spike.New() // Use the default Workload API Socket
	if connErr != nil {
		fmt.Println("Error connecting to SPIKE Nexus:", connErr.Error())
		return
	}

	fmt.Println("Connected to SPIKE Nexus.")

	// https://pkg.go.dev/github.com/spiffe/spike-sdk-go/api#Close
	defer func() {
		// Close the connection when done
		closeErr := api.Close()
		if closeErr != nil {
			fmt.Println("Error closing connection:", closeErr.Error())
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
		fmt.Println("Error writing secret:", putErr.Error())
		return
	}

	// Read the Secret
	// https://pkg.go.dev/github.com/spiffe/spike-sdk-go/api#GetSecret
	secret, getErr := api.GetSecret(ctx, path)
	if getErr != nil {
		fmt.Println("Error reading secret:", getErr.Error())
		return
	}

	if secret == nil {
		fmt.Println("Secret not found.")
		return
	}

	fmt.Println("Secret found:")
	for k, v := range secret.Data {
		fmt.Printf("%s: %s\n", k, v)
	}
}
