//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

// Package handle provides the implementation for registering and managing
// HTTP route handlers for the application. This package is responsible for
// setting up the route configuration during the application's initialization
// phase.
package handle

import (
	state "github.com/spiffe/spike/app/keeper/internal/route/base"

	"github.com/spiffe/spike/internal/net"
)

// InitializeRoutes registers the main HTTP route handler for the application.
// It sets up a single catch-all route "/" that forwards all requests to the
// route.Route handler.
//
// This function should be called during application startup, before starting
// the HTTP server.
func InitializeRoutes() {
	net.HandleRoute(state.Route)
}
