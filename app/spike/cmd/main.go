//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"

	"github.com/spiffe/spike/app/spike/internal/cmd"
	"github.com/spiffe/spike/internal/config"
	"github.com/spiffe/spike/internal/log"
	"github.com/spiffe/spike/internal/spiffe"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	source, spiffeid, err := spiffe.AppSpiffeSource(ctx)
	if err != nil {
		log.Fatal(err.Error())
	}
	defer spiffe.CloseSource(source)

	if !config.IsPilot(spiffeid) {
		log.FatalF("SPIFFE ID %s is not valid.\n", spiffeid)
	}

	cmd.Initialize(source)
	cmd.Execute()
}
