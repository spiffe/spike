//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package route

import (
	"net/http"
	"time"

	"github.com/spiffe/spike/internal/log"
	"github.com/spiffe/spike/internal/net"
)

func logAndRoute(entry log.AuditEntry, h handler) handler {
	log.Audit(entry)
	return h
}

func factory(p, a, m string) handler {
	now := time.Now()
	entry := log.AuditEntry{
		Timestamp: now,
		UserId:    "TBD",
		Action:    "",
		Resource:  p,
		SessionID: "",
	}

	switch {
	case m == http.MethodPost && a == "" && p == urlKeep:
		entry.Action = "create"
		return logAndRoute(entry, routeKeep)
	case m == http.MethodPost && a == "read" && p == urlKeep:
		entry.Action = "read"
		return logAndRoute(entry, routeShow)
	// Fallback route.
	default:
		entry.Action = "fallback"
		return logAndRoute(entry, net.Fallback)
	}
}
