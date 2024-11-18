//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"fmt"

	"github.com/spiffe/spike/app/demo/internal/spiffe"
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
	// We cannot have assumptions about how the app SPIFFE ID is formatted.
	// We need to do some form of registration on SPIKE Nexus to validate the
	// SPIFFE ID.
	fmt.Println("SPIFFE ID:", spiffeid)

	fmt.Println("Demo app authenticated")
}
