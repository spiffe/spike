//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package main

import (
	"crypto/fips140"
	"flag"
	"fmt"
	"os"

	"github.com/spiffe/spike-sdk-go/log"
	"github.com/spiffe/spike-sdk-go/spiffe"
	svid "github.com/spiffe/spike-sdk-go/spiffeid"
	"github.com/spiffe/spike/internal/config"

	"github.com/spiffe/spike/app/bootstrap/internal/env"
	"github.com/spiffe/spike/app/bootstrap/internal/net"
	"github.com/spiffe/spike/app/bootstrap/internal/state"
	"github.com/spiffe/spike/app/bootstrap/internal/url"
)

func main() {
	const fName = "boostrap.main"

	log.Log().Info(fName, "message", "Starting SPIKE bootstrap...", "version", config.BootstrapVersion)

	init := flag.Bool("init", false, "Initialize the bootstrap module")
	flag.Parse()
	if !*init {
		fmt.Println("")
		fmt.Println("Usage: bootstrap -init")
		fmt.Println("")
		os.Exit(1)
		return
	}

	// Check if we should skip bootstrap (set by init container)
	if _, err := os.Stat("/shared/skip-bootstrap"); err == nil {
		log.Log().Info(fName,
			"message", "Bootstrap already completed previously. Skipping.",
		)
		fmt.Println("Bootstrap already completed previously. Exiting.")
		os.Exit(0)
		return
	}

	src := net.Source()
	defer spiffe.CloseSource(src)
	sv, err := src.GetX509SVID()
	if err != nil {
		log.FatalLn(fName,
			"message", "Failed to get X.509 SVID",
			"err", err.Error())
		os.Exit(1)
		return
	}

	if !svid.IsBootstrap(env.TrustRoot(), sv.ID.String()) {
		log.Log().Error(
			"Authenticate: You need a 'boostrap' SPIFFE ID to use this command.",
		)
		os.Exit(1)
		return
	}

	log.Log().Info(
		fName, "FIPS 140.3 enabled", fips140.Enabled(),
	)
	log.Log().Info(
		fName, "message", "Sending shards to SPIKE Keeper instances...",
	)
	for keeperID, keeperAPIRoot := range env.Keepers() {
		log.Log().Info(fName, "keeper ID", keeperID)
		net.Post(
			net.MTLSClient(src),
			url.KeeperEndpoint(keeperAPIRoot),
			net.Payload(
				state.KeeperShare(
					state.RootShares(), keeperID),
				keeperID,
			),
			keeperID,
		)
	}

	log.Log().Info(fName, "message", "Sent shards to SPIKE Keeper instances.")
	fmt.Println("Bootstrap completed successfully!")
	os.Exit(0)
}
