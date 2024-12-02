//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"fmt"

	spike "github.com/spiffe/spike-sdk-go/api"
	"github.com/spiffe/spike-sdk-go/spiffe"
)

func main() {
	// Create a context.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize the SPIFFE endpoint socket.
	defaultEndpointSocket := spiffe.EndpointSocket()

	// Initialize the SPIFFE source.
	source, spiffeid, err := spiffe.Source(ctx, defaultEndpointSocket)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	// Close the SPIFFE source when done.
	defer spiffe.CloseSource(source)

	fmt.Println("SPIFFE ID:", spiffeid)

	//
	// Retrieve a secret using SPIKE SDK.
	//

	path := "/tenants/demo/db/creds"
	version := 0

	secret, err := spike.GetSecret(source, path, version)
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
