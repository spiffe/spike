//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package net

import (
	"github.com/spiffe/go-spiffe/v2/workloadapi"
	"github.com/spiffe/spike-sdk-go/config/env"
	sdkErrors "github.com/spiffe/spike-sdk-go/errors"
	"github.com/spiffe/spike-sdk-go/log"
	"github.com/spiffe/spike-sdk-go/net"

	http "github.com/spiffe/spike/app/nexus/internal/route/base"
	routing "github.com/spiffe/spike/internal/net"
)

// Serve initializes and starts a TLS-secured HTTP server for the given
// application.
//
// Serve uses the provided X509Source for TLS authentication and configures the
// server with the specified HTTP routes. It will listen on the port specified
// by the TLS port environment variable. If the server fails to start, it logs a
// fatal error and terminates the application.
//
// Parameters:
//   - appName: A string identifier for the application, used in error messages
//   - source: An X509Source that provides TLS certificates for the server. Can
//     be nil, but if nil at startup, the function will crash with log.FatalErr
//     since the server cannot operate without mTLS credentials and there is no
//     retry mechanism for server initialization. This fail-fast behavior makes
//     configuration or initialization problems immediately obvious to
//     operators.
//
// The function does not return unless an error occurs, in which case it calls
// log.FatalErr and terminates the program.
func Serve(appName string, source *workloadapi.X509Source) {
	// Fail-fast if the source is nil: server cannot operate without mTLS
	if source == nil {
		log.FatalErr(appName, *sdkErrors.ErrSPIFFENilX509Source.Clone())
	}

	if serveErr := net.Serve(
		source,
		func() { routing.HandleRoute(http.Route) },
		env.NexusTLSPortVal(),
	); serveErr != nil {
		log.FatalErr(appName, *serveErr)
	}
}
