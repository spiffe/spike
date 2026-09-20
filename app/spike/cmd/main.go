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
	"github.com/spiffe/spike/app/spike/internal/logfile"
)

const appName = "SPIKE"

func main() {
	// Route the SDK logger to the Pilot diagnostics log before any other
	// call can bind it to stdout: command output must stay clean.
	if routeErr := logfile.Route(); routeErr != nil {
		if _, err := fmt.Fprintf(os.Stderr,
			"Error: cannot open the diagnostics log %s: %v\n",
			logfile.Path(), routeErr,
		); err != nil {
			log.FatalLn(appName, "message", "failed to write to stderr",
				"err", err.Error())
		}
		log.FatalErr(appName, *routeErr)
	}

	errMem := mem.Lock()
	if errMem != nil {
		if env.ShowMemoryWarningVal() {
			if _, err := fmt.Fprintln(os.Stderr, `
Memory locking is not available.
Consider disabling swap to enhance security.
 `); err != nil {
				// The Pilot cannot reach its own stderr; treat this as a
				// broken environment rather than continuing silently.
				log.FatalLn(appName, "message", "failed to write to stderr",
					"err", err.Error())
			}
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	source, SPIFFEID, err := spiffe.Source(ctx, spiffe.EndpointSocket())
	if err != nil {
		failErr := sdkErrors.ErrStateInitializationFailed.Wrap(err)
		log.FatalErr(appName, *failErr)
	}
	defer func() {
		if closeErr := spiffe.CloseSource(source); closeErr != nil {
			warnErr := sdkErrors.ErrSPIFFEFailedToCloseX509Source.Wrap(closeErr)
			log.WarnErr(appName, *warnErr)
		}
	}()

	cmd.Initialize(source, SPIFFEID)
	cmd.Execute()
}
