//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"fmt"
	"os"

	"github.com/spiffe/spike-sdk-go/config/env"
	sdkErrors "github.com/spiffe/spike-sdk-go/errors"
	"github.com/spiffe/spike-sdk-go/log"
	"github.com/spiffe/spike-sdk-go/security/mem"
	"github.com/spiffe/spike-sdk-go/spiffe"

	"github.com/spiffe/spike/app/spike/internal/cmd"
)

const appName = "SPIKE"

func main() {
	if !mem.Lock() {
		if env.ShowMemoryWarningVal() {
			if _, err := fmt.Fprintln(os.Stderr, `
Memory locking is not available.
Consider disabling swap to enhance security.
 `); err != nil {
				fmt.Println("failed to write to stderr: ", err.Error())
			}
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	source, SPIFFEID, err := spiffe.Source(ctx, spiffe.EndpointSocket())
	if err != nil {
		failErr := sdkErrors.ErrInitializationFailed.Wrap(err)
		log.FatalErr(appName, *failErr)
	}
	defer spiffe.CloseSource(source)

	cmd.Initialize(source, SPIFFEID)
	cmd.Execute()
}
