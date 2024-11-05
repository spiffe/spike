//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package handle

import (
	"net/http"

	"github.com/spiffe/spike/app/nexus/internal/route"
)

// InitializeRoutes registers the main HTTP route handler for the application.
// It sets up a single catch-all route "/" that forwards all requests to the
// route.Route handler.
//
// This function should be called during application startup, before starting
// the HTTP server.
func InitializeRoutes() {
	http.HandleFunc("/", route.Route)
}
