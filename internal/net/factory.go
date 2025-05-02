//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package net

import (
	"net/http"

	"github.com/spiffe/spike-sdk-go/api/url"

	"github.com/spiffe/spike/internal/log"
)

// RouteFactory creates HTTP route handlers for API endpoints using a generic
// switching function. It enforces POST-only methods per ADR-0012 and logs
// route creation details.
//
// Type Parameters:
//   - ApiAction: Type representing the API action to be handled
//
// Parameters:
//   - p: API URL for the route
//   - a: API action instance
//   - m: HTTP method
//   - switchyard: Function that returns an appropriate handler based on
//     action and URL
//
// Returns:
//   - Handler: Route handler function or Fallback for non-POST methods
func RouteFactory[ApiAction any](p url.ApiUrl, a ApiAction, m string,
	switchyard func(a ApiAction, p url.ApiUrl) Handler) Handler {
	log.Log().Info("route.factory", "path", p, "action", a, "method", m)

	// We only accept POST requests -- See ADR-0012.
	if m != http.MethodPost {
		return Fallback
	}

	return switchyard(a, p)
}
