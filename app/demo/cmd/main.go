//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"fmt"

	"github.com/spiffe/spike/app/demo/spike"
	"github.com/spiffe/spike/pkg/spiffe"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	source, spiffeid, err := spiffe.AppSpiffeSource(ctx)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	defer spiffe.CloseSource(source)

	fmt.Println("Demo app initialized")
	fmt.Println("SPIFFE ID:", spiffeid)

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
