//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package route

import (
	"fmt"
	"github.com/spiffe/spike/internal/log"
	"net/http"
)

func Route(
	w http.ResponseWriter, r *http.Request, audit *log.AuditEntry,
) error {
	fmt.Println(">>>>>>>>>>>>>>>>>>", "path", r.URL.Path, "action", r.URL.Query().Get("action"), "method", r.Method)
	return factory(
		r.URL.Path, r.URL.Query().Get("action"), r.Method,
	)(w, r, audit)
}
