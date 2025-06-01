//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"fmt"
	"github.com/spiffe/spike-sdk-go/security/mem"
	"github.com/spiffe/spike-sdk-go/spiffe"

	"github.com/spiffe/spike/app/spike/internal/cmd"
	"github.com/spiffe/spike/internal/log"
)

func main() {
	if !mem.Lock() {
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "Memory locking is not available.\n")
		fmt.Fprintf(os.Stderr, "Consider disabling swap to enhance security.\n")
		fmt.Fprintf(os.Stderr, "\n")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	source, spiffeId, err := spiffe.Source(ctx, spiffe.EndpointSocket())
	if err != nil {
		log.Fatal(err.Error())
	}
	defer spiffe.CloseSource(source)

	cmd.Initialize(source, spiffeId)
	cmd.Execute()
}
