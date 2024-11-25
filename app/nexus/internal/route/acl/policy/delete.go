//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package policy

import (
	"errors"
	"net/http"

	state "github.com/spiffe/spike/app/nexus/internal/state/base"
	"github.com/spiffe/spike/internal/entity/v1/reqres"
	"github.com/spiffe/spike/internal/log"
	"github.com/spiffe/spike/internal/net"
)

func RouteDeletePolicy(
	w http.ResponseWriter, r *http.Request, audit *log.AuditEntry,
) error {
	log.Log().Info("routeDeletePolicy", "method", r.Method, "path", r.URL.Path,
		"query", r.URL.RawQuery)
	audit.Action = log.AuditDelete

	requestBody := net.ReadRequestBody(w, r)
	if requestBody == nil {
		return errors.New("failed to read request body")
	}

	request := net.HandleRequest[
		reqres.PolicyDeleteRequest, reqres.PolicyDeleteResponse](
		requestBody, w,
		reqres.PolicyDeleteResponse{Err: reqres.ErrBadInput},
	)
	if request == nil {
		return errors.New("failed to parse request body")
	}

	policyId := request.Id

	err := state.DeletePolicy(policyId)
	if err != nil {
		log.Log().Info("routeDeletePolicy",
			"msg", "Failed to delete policy", "err", err)

		responseBody := net.MarshalBody(reqres.PolicyDeleteResponse{
			Err: "Internal server error",
		}, w)
		if responseBody == nil {
			return errors.New("failed to marshal response body")
		}

		net.Respond(http.StatusInternalServerError, responseBody, w)
		log.Log().Info("routeDeletePolicy", "msg", "internal server error")
		return err
	}

	responseBody := net.MarshalBody(reqres.PolicyDeleteResponse{}, w)
	if responseBody == nil {
		return errors.New("failed to marshal response body")
	}

	net.Respond(http.StatusOK, responseBody, w)
	log.Log().Info("routeDeletePolicy", "msg", "OK")
	return nil
}
