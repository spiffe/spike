//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"fmt"

	"github.com/spiffe/spike-sdk-go/security/mem"
	"github.com/spiffe/spike-sdk-go/spiffe"
	"github.com/spiffe/spike-sdk-go/spiffeid"

	"github.com/spiffe/spike/app/nexus/internal/env"
	"github.com/spiffe/spike/app/nexus/internal/initialization"
	"github.com/spiffe/spike/app/nexus/internal/net"
	"github.com/spiffe/spike/internal/config"
	"github.com/spiffe/spike/internal/log"
)

const appName = "SPIKE Nexus"

func main() {
	if env.BannerEnabled() {
		fmt.Printf(`
   \\ SPIKE: Secure your secrets with SPIFFE.
 \\\\\ Copyright 2024-present SPIKE contributors.
\\\\\\\ SPDX-License-Identifier: Apache-2.0`+"\n\n"+
			"%s v%s. | LOG LEVEL: %s\n\n",
			appName, config.SpikeNexusVersion, log.Level(),
		)
	}

	if mem.Lock() {
		log.Log().Info(appName, "msg", "Successfully locked memory.")
	} else {
		log.Log().Info(appName, "msg", "Memory is not locked. Please disable swap.")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log.Log().Info(appName, "msg", "SPIFFE Trust Domain: "+env.TrustRoot())

	fmt.Println("before trying to get source...")
	source, selfSpiffeid, err := spiffe.Source(ctx, spiffe.EndpointSocket())
	if err != nil {
		log.Fatal(err.Error())
	}
	defer spiffe.CloseSource(source)

	log.Log().Info(appName, "msg", "self.spiffeid: "+selfSpiffeid)

	// I should be Nexus.
	if !spiffeid.IsNexus(env.TrustRoot(), selfSpiffeid) {
		log.FatalF("Authenticate: SPIFFE ID %s is not valid.\n", selfSpiffeid)
	}

	initialization.Initialize(source)

	log.Log().Info(appName, "msg", fmt.Sprintf(
		"Started service: %s v%s",
		appName, config.SpikeNexusVersion),
	)

	// Start the server:
	net.Serve(appName, source)
}
