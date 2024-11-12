//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package route

import (
	"net/http"
	"time"

	"github.com/spiffe/spike/internal/log"
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
	case m == http.MethodPost && a == "admin" && p == urlLogin:
		entry.Action = "admin-login"
		return logAndRoute(entry, routeAdminLogin)
	case m == http.MethodPost && a == "" && p == urlInit:
		entry.Action = "init"
		return logAndRoute(entry, routeInit)
	case m == http.MethodPost && a == "check" && p == urlInit:
		entry.Action = "check"
		return logAndRoute(entry, routeInitCheck)
	case m == http.MethodPost && a == "" && p == urlSecrets:
		entry.Action = "create"
		return logAndRoute(entry, routePutSecret)
	case m == http.MethodPost && a == "get" && p == urlSecrets:
		entry.Action = "read"
		return logAndRoute(entry, routeGetSecret)
	case m == http.MethodPost && a == "delete" && p == urlSecrets:
		entry.Action = "delete"
		return logAndRoute(entry, routeDeleteSecret)
	case m == http.MethodPost && a == "undelete" && p == urlSecrets:
		entry.Action = "undelete"
		return logAndRoute(entry, routeUndeleteSecret)
	case m == http.MethodPost && a == "list" && p == urlSecrets:
		entry.Action = "list"
		return logAndRoute(entry, routeListPaths)
	// Fallback route.
	default:
		entry.Action = "fallback"
		return logAndRoute(entry, routeFallback)
	}
}
