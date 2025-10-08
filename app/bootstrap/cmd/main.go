//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"crypto/fips140"
	"flag"
	"fmt"
	"time"

	"github.com/spiffe/spike-sdk-go/log"
	"github.com/spiffe/spike-sdk-go/retry"
	"github.com/spiffe/spike-sdk-go/spiffe"
	svid "github.com/spiffe/spike-sdk-go/spiffeid"

	"github.com/spiffe/spike/app/bootstrap/internal/env"
	"github.com/spiffe/spike/app/bootstrap/internal/lifecycle"
	"github.com/spiffe/spike/app/bootstrap/internal/net"
	"github.com/spiffe/spike/app/bootstrap/internal/state"
	"github.com/spiffe/spike/app/bootstrap/internal/url"
	"github.com/spiffe/spike/internal/config"
)

func main() {
	const fName = "bootstrap.main"

	log.Log().Info(fName, "message",
		"Starting SPIKE bootstrap...",
		"version", config.BootstrapVersion,
	)

	init := flag.Bool("init", false, "Initialize the bootstrap module")
	flag.Parse()
	if !*init {
		fmt.Println("")
		fmt.Println("Usage: bootstrap -init")
		fmt.Println("")
		log.FatalLn(fName, "message", "Invalid command line arguments")
		return
	}

	skip := !lifecycle.ShouldBootstrap() // Kubernetes or bare-metal check.
	if skip {
		log.Log().Info(fName,
			"message", "Skipping bootstrap.",
		)
		fmt.Println("Bootstrap skipped. Check the logs for more information.")
		return
	}

	src := net.Source()
	defer spiffe.CloseSource(src)
	sv, err := src.GetX509SVID()
	if err != nil {
		log.FatalLn(fName,
			"message", "Failed to get X.509 SVID",
			"err", err.Error())
		log.FatalLn(fName, "message", "Failed to acquire SVID")
		return
	}

	if !svid.IsBootstrap(sv.ID.String()) {
		log.Log().Error(
			"Authenticate: You need a 'bootstrap' SPIFFE ID to use this command.",
		)
		log.FatalLn(fName, "message", "Command not authorized")
		return
	}

	log.Log().Info(
		fName, "FIPS 140.3 enabled", fips140.Enabled(),
	)

	log.Log().Info(
		fName, "message", "Sending shards to SPIKE Keeper instances...",
	)

	ctx := context.Background()

	for keeperID, keeperAPIRoot := range env.Keepers() {
		log.Log().Info(fName, "keeper ID", keeperID)

		_, err := retry.Do(ctx, func() (bool, error) {
			log.Log().Info(fName, "message", "retry:"+time.Now().String())

			err := net.Post(
				net.MTLSClient(src),
				url.KeeperContributeEndpoint(keeperAPIRoot),
				net.Payload(
					state.KeeperShare(
						state.RootShares(), keeperID),
					keeperID,
				),
				keeperID,
			)
			if err != nil {
				log.Log().Warn(fName, "message", "Failed to send shard. Will retry.")
				return false, err
			}

			log.Log().Info(fName, "message", "Shard sent successfully.")
			return true, nil
		},
			retry.WithBackOffOptions(
				retry.WithMaxInterval(60*time.Second), // TODO: to env vars.
				retry.WithMaxElapsedTime(0),           // Retry forever.
			),
		)

		// This should never happen since the above loop retries forever:
		if err != nil {
			log.FatalLn(fName, "message", "Initialization failed", "err", err)
		}
	}

	// TODO:
	// 1. Create a random text
	// 2. Encrypt it with the root key
	// 3. Send it to SPIKE Nexus
	// 4. SPIKE Nexus will decrypt it and send its hash back.
	// 5. Verify the hash to ensure that SPIKE Nexus has initialized.

	log.Log().Info(fName, "message", "Sent shards to SPIKE Keeper instances.")

	// Mark completion in Kubernetes
	if err := lifecycle.MarkBootstrapComplete(); err != nil {
		// Log but don't fail - bootstrap itself succeeded
		log.Log().Warn(fName, "message",
			"Could not mark bootstrap complete in ConfigMap", "err", err.Error())
	}

	fmt.Println("Bootstrap completed successfully!")
}
