//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"fmt"

	"github.com/spiffe/spike-sdk-go/net"
	"github.com/spiffe/spike-sdk-go/spiffe"

	"github.com/spiffe/spike/app/nexus/internal/env"
	"github.com/spiffe/spike/app/nexus/internal/initialization"
	"github.com/spiffe/spike/app/nexus/internal/route/handle"
	"github.com/spiffe/spike/app/nexus/internal/trust"
	"github.com/spiffe/spike/internal/config"
	"github.com/spiffe/spike/internal/log"
)

const appName = "SPIKE Nexus"

// TODO:
// Document this recovery flow in the website.
//
// TODO: create a demo video recording.
//
// 1. Superadmin runs `spike recover` on SPIKE Pilot.
// 2. System tells the superadmin to assume a role by running a script on the
//    SPIRE Server.
//    "spiffe://" + trustRoot() + "/spike/pilot/role/recover"
//    the CLI can even print out the command to run.
//    this SVID can only do the recovery function, cannot use any other API.
// 3. Superadmin runs the command on the SPIRE Server assuming the role.
// 4. Superadmin runs `spike recover` on SPIKE Pilot again.
// 5. Two shards are saved on the file system and instructions printed on how
//    to use them, wipe them out cleanly etc.
// 6. Superadmin saves these files securely, in an encrypted medium.
// 8. Now we are ready for doomsday.
// -------------------------------
// 1. Assume running SPIKE Nexus with fully hydrated SPIKE Keepers.
// 2. Stop SPIKE Nexus.
// 3. Restart all SPIKE Keepers.
// 4. Stop SPIKE Nexus.
// 5. ASSERT that SPIKE Nexus falls into the RecoverBackingStoreUsingKeeperShards loop.
//    (i.e. recovery mode)
//
// 6. Super adming runs `spike restore` and guided to change the SVID again.
// 7. Superadmin runs a script to change Pilot SVID to
//    "spiffe://" + trustRoot() + "/spike/pilot/role/restore"
//    this SVID can only do restore functionality.
//    these SVIDs will also be good for auditing purposes.
// 8. Superadmin runs `spike restore` on SPIKE Pilot again.
// 9. Superadmin inputs one shard.
// 10. Superadmin inputs the second shard.
// 11. The system is restored.
// 12. ASSERT that we are out of the RecoverBackingStoreUsingKeeperShards loop.
// 13. CLI also notifies admin to change the SVID back to the superuser svid.
// 14. Everyone is happy.
// ------
// What we need on the CLI:
// `spike recover` and `spike restore`
// What we need on the Nexus API:
// /recover and /restore
//------------------------------------------------------------------------------
//------------------------------------------------------------------------------
// Below is old thought process, leaving for reference:
//
// 1. Ensure that SPIKE Nexus can successfully recover.
// 2. Have a /health API for keepers. If SPIKE Nexus cannot get a healthy
//      response from a keeper, it should run a background process to re-send
//      shards. The /health API should be more frequent than the periodic shard
//      syncing. The health API will report whether the keeper has a shard too
//      so that nexus can decide to re-seed it or not.
//
// 3. Doomsday recovery:
//    3.1. Have a /recover endpoint that dumps two shards as a response.
//    3.2. Have SPIKE Pilot save those shards in two separate files under ~/.spike
//    3.3. Have a /restore endpoint on SPIKE Nexus that accepts one shard at
//     a time.
//    3.4. When restore has enough shards, it recomputes the root key, updates
//     internal state, starts a new bootstrapping flow.
//    other notes:
//      - Keepers are designed to not know anything about Nexus or themselves
//        this means, we can send them data, but they wouldn't know about their
//        actual state (i.e. whether they are at day zero and that's why they
//        don't have a shard; or they crashed and they don't have a shard.
//        only thing they have is a health endpoint. and nexus is responsible
//        to reseed them, if they report unhealthy.

func main() {
	log.Log().Info(appName, "msg", appName, "version", config.NexusVersion)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	source, spiffeid, err := spiffe.Source(ctx, spiffe.EndpointSocket())
	if err != nil {
		log.Fatal(err.Error())
	}
	defer spiffe.CloseSource(source)

	trust.Authenticate(spiffeid)

	initialization.Initialize(source)

	log.Log().Info(appName, "msg", fmt.Sprintf(
		"Started service: %s v%s",
		appName, config.NexusVersion),
	)

	if err := net.Serve(
		source, handle.InitializeRoutes,
		env.TlsPort(),
	); err != nil {
		log.FatalF("%s: Failed to serve: %s\n", appName, err.Error())
	}
}
