//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"fmt"

	"github.com/spiffe/spike-sdk-go/security/mem"
	"github.com/spiffe/spike-sdk-go/spiffe"
	"github.com/spiffe/spike-sdk-go/spiffeid"

	"github.com/spiffe/spike/app/keeper/internal/env"
	"github.com/spiffe/spike/app/keeper/internal/net"
	"github.com/spiffe/spike/internal/config"
	"github.com/spiffe/spike/internal/log"
)

const appName = "SPIKE Keeper"

func main() {
	if env.BannerEnabled() {
		fmt.Printf(`
   \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
 \\\\\ Copyright 2024-present SPIKE contributors.
\\\\\\\ SPDX-License-Identifier: Apache-2.0`+"\n\n"+
			"%s v%s. | LOG LEVEL: %s\n\n",
			appName, config.SpikeKeeperVersion, log.Level(),
		)
	}

	if mem.Lock() {
		log.Log().Info(appName, "msg", "Successfully locked memory.")
	} else {
		log.Log().Info(appName, "msg", "Memory is not locked. Please disable swap.")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	source, selfSpiffeid, err := spiffe.Source(ctx, spiffe.EndpointSocket())
	if err != nil {
		log.FatalLn(err.Error())
	}
	defer spiffe.CloseSource(source)

	// I should be a SPIKE Keeper.
	if !spiffeid.IsKeeper(env.TrustRoot(), selfSpiffeid) {
		log.FatalF("Authenticate: SPIFFE ID %s is not valid.\n", selfSpiffeid)
	}

	log.Log().Info(
		appName, "msg",
		fmt.Sprintf("Started service: %s v%s", appName, config.SpikeKeeperVersion),
	)

	// Serve the app:
	net.Serve(appName, source)
}
