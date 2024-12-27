//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package store

import (
	"net/http"

	"github.com/spiffe/spike/internal/log"
)

func RouteStatus(w http.ResponseWriter, r *http.Request, audit *log.AuditEntry) error {
	// TODO: implement me.
	// TODO: maybe nexus would want to call me before querying for shards.

	return nil
}
