//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"crypto/fips140"
	"flag"
	"fmt"

	spike "github.com/spiffe/spike-sdk-go/api"
	"github.com/spiffe/spike-sdk-go/log"

	"github.com/spiffe/spike/app/bootstrap/internal/lifecycle"
	"github.com/spiffe/spike/app/bootstrap/internal/net"
	"github.com/spiffe/spike/internal/config"
)

const appName = "SPIKE Bootstrap"

func main() {
	log.Log().Info(
		appName,
		"message", "starting SPIKE bootstrap...",
		"version", config.BootstrapVersion,
	)

	init := flag.Bool("init", false, "Initialize the bootstrap module")
	flag.Parse()
	if !*init {
		fmt.Println("")
		fmt.Println("Usage: bootstrap -init")
		fmt.Println("")
		log.FatalLn(appName, "message", "Invalid command line arguments")
		return
	}

	skip := !lifecycle.ShouldBootstrap() // Kubernetes or bare-metal check.
	if skip {
		log.Log().Info(appName, "message", "skipping bootstrap")
		fmt.Println("Bootstrap skipped. Check the logs for more information.")
		return
	}

	log.Log().Info(
		appName, "message", "FIPS 140.3 Status", "enabled", fips140.Enabled(),
	)

	// Panics if it cannot acquire the source.
	src := net.AcquireSource()

	log.Log().Info(
		appName, "message", "sending shards to SPIKE Keeper instances",
	)

	api := spike.NewWithSource(src)
	defer api.Close()

	ctx := context.Background()

	// Broadcast shards to the SPIKE keepers until all shards are
	// dispatched successfully.
	net.BroadcastKeepers(ctx, api)

	log.Log().Info(appName, "message", "sent shards to SPIKE Keeper instances")

	// Verify that SPIKE Nexus has been properly initialized by sending an
	// encrypted payload and verifying the hash of the decrypted plaintext.
	// Retries verification until successful.
	net.VerifyInitialization(ctx, api)

	// Bootstrap verification is complete. Mark the bootstrap as "done".

	// Mark completion in Kubernetes
	if err := lifecycle.MarkBootstrapComplete(); err != nil {
		// Log but don't fail - bootstrap itself succeeded
		log.Log().Warn(
			appName,
			"message", "could not mark bootstrap complete in ConfigMap",
			"err", err.Error(),
		)
	}

	fmt.Println("bootstrap completed successfully")
}
