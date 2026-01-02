//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"crypto/fips140"
	"flag"
	"time"

	spike "github.com/spiffe/spike-sdk-go/api"
	"github.com/spiffe/spike-sdk-go/config/env"
	sdkErrors "github.com/spiffe/spike-sdk-go/errors"
	"github.com/spiffe/spike-sdk-go/log"
	"github.com/spiffe/spike-sdk-go/security/mem"
	"github.com/spiffe/spike/app/bootstrap/internal/state"

	"github.com/spiffe/spike/app/bootstrap/internal/lifecycle"
	"github.com/spiffe/spike/app/bootstrap/internal/net"
	"github.com/spiffe/spike/internal/config"
)

const appName = "SPIKE Bootstrap"

func main() {
	log.Info(
		appName,
		"message", "starting",
		"version", config.BootstrapVersion,
	)

	// Hard timeout for the entire bootstrap process.
	// A value of 0 means no timeout (infinite).
	bootstrapTimeout := env.BootstrapTimeoutVal()
	if bootstrapTimeout > 0 {
		log.Debug(
			appName,
			"message", "bootstrap timeout configured",
			"timeout", bootstrapTimeout.String(),
		)
		time.AfterFunc(bootstrapTimeout, func() {
			log.FatalLn(
				appName,
				"message", "bootstrap timeout exceeded, terminating",
				"timeout", bootstrapTimeout.String(),
			)
		})
	}

	init := flag.Bool("init", false, "Initialize the bootstrap module")
	flag.Parse()
	if !*init {
		failErr := *sdkErrors.ErrDataInvalidInput.Clone()
		failErr.Msg = "invalid command line arguments: usage: boostrap -init"
		log.FatalErr(appName, failErr)
		return
	}

	//  0. Skip bootstrap for Lite backend and In-Memory backend
	//  1. Else, if SPIKE_BOOTSTRAP_FORCE="true", always proceed (return true)
	//  2. In bare-metal environments (non-Kubernetes), always proceed
	//  3. In Kubernetes environments, check the "spike-bootstrap-state"
	//     ConfigMap:
	//     - If ConfigMap exists and bootstrap-completed="true", skip bootstrap
	//     - Otherwise, proceed with bootstrap
	skip := !lifecycle.ShouldBootstrap()
	if skip {
		log.Warn(appName, "message", "skipping bootstrap")
		return
	}

	log.Info(
		appName,
		"message", "FIPS 140.3 Status",
		"enabled", fips140.Enabled(),
	)

	// Panics if it cannot acquire the source.
	src := net.AcquireSource()

	log.Info(appName, "message", "sending shards to SPIKE Keeper instances")

	api := spike.NewWithSource(src)
	defer func() {
		closeErr := api.Close()
		if closeErr == nil {
			return
		}
		warnErr := sdkErrors.ErrFSStreamCloseFailed.Wrap(closeErr)
		warnErr.Msg = "failed to close SPIKE API client"
		log.WarnErr(appName, *warnErr)
	}()

	ctx := context.Background()

	// Broadcast shards to the SPIKE keepers until all shards are
	// dispatched successfully.
	net.BroadcastKeepers(ctx, api)

	log.Info(appName, "message", "sent shards to SPIKE Keeper instances")

	// Verify that SPIKE Nexus has been properly initialized by sending an
	// encrypted payload and verifying the hash of the decrypted plaintext.
	// Retries verification until successful.
	net.VerifyInitialization(ctx, api)

	// Clear the seed after use.
	state.LockRootKeySeed()
	defer state.UnlockRootKeySeed()
	mem.ClearRawBytes(state.RootKeySeedNoLock())

	// Bootstrap verification is complete. Mark the bootstrap as "done".

	// Mark completion in Kubernetes
	if err := lifecycle.MarkBootstrapComplete(); err != nil {
		warnErr := sdkErrors.ErrK8sReconciliationFailed.Wrap(err)
		warnErr.Msg = "failed to mark bootstrap complete in ConfigMap"
		log.WarnErr(appName, *warnErr)
	}

	log.Info("bootstrap completed successfully")
}
