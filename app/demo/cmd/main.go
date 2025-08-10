//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"

	spike "github.com/spiffe/spike-sdk-go/api"
)

func main() {
	// https://pkg.go.dev/github.com/spiffe/spike-sdk-go/api#New
	api := spike.New() // Use the default Workload API Socket

	// https://pkg.go.dev/github.com/spiffe/spike-sdk-go/api#Close
	defer api.Close() // Close the connection when done

	path := "^tenants/demo/db/creds"

	// Create a Secret
	// https://pkg.go.dev/github.com/spiffe/spike-sdk-go/api#PutSecret
	err := api.PutSecret(path, map[string]string{
		"username": "SPIKE",
		"password": "SPIKE_Rocks",
	})
	if err != nil {
		fmt.Println("Error writing secret:", err.Error())
		return
	}

	// Read the Secret
	// https://pkg.go.dev/github.com/spiffe/spike-sdk-go/api#GetSecret
	secret, err := api.GetSecret(path)
	if err != nil {
		fmt.Println("Error reading secret:", err.Error())
		return
	}

	if secret == nil {
		fmt.Println("Secret not found.")
		return
	}

	fmt.Println("Secret found:")

	data := secret.Data
	for k, v := range data {
		fmt.Printf("%s: %s\n", k, v)
	}
}
