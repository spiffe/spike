//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/spiffe/spike-sdk-go/spiffe"

	"github.com/spiffe/spike/app/bootstrap/internal/env"
	"github.com/spiffe/spike/app/bootstrap/internal/net"
	"github.com/spiffe/spike/app/bootstrap/internal/state"
	"github.com/spiffe/spike/app/bootstrap/internal/url"
)

func main() {
	init := flag.Bool("init", false, "Initialize the bootstrap module")
	flag.Parse()
	if !*init {
		fmt.Println("")
		fmt.Println("Usage: bootstrap -init")
		fmt.Println("")
		os.Exit(1)
	}

	src := net.Source()
	defer spiffe.CloseSource(src)

	for keeperID, keeperAPIRoot := range env.Keepers() {
		net.Post(
			net.MTLSClient(src),
			url.KeeperEndpoint(keeperAPIRoot),
			net.Payload(state.KeeperShare(state.RootShares(), keeperID), keeperID),
			keeperID,
		)
	}
}
