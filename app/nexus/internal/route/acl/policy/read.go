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

// TODO: the request and response can be generic and part of Route's signature.

func RouteGetPolicy(
	w http.ResponseWriter, r *http.Request, audit *log.AuditEntry,
) error {
	log.Log().Info("routeGetPolicy", "method", r.Method, "path", r.URL.Path,
		"query", r.URL.Query())
	audit.Action = log.AuditRead

	requestBody := net.ReadRequestBody(w, r)
	if requestBody == nil {
		return errors.New("failed to read request body")
	}

	request := net.HandleRequest[
		reqres.PolicyReadRequest, reqres.PolicyReadResponse](
		requestBody, w,
		reqres.PolicyReadResponse{Err: reqres.ErrBadInput},
	)
	if request == nil {
		return errors.New("failed to parse request body")
	}

	policyId := request.Id

	policy, err := state.GetPolicy(policyId)
	if err == nil {
		log.Log().Info("routeGetPolicy", "msg", "Policy found")
	} else if errors.Is(err, state.ErrPolicyNotFound) {
		log.Log().Info("routeGetPolicy", "msg", "Policy not found")

		res := reqres.PolicyReadResponse{Err: reqres.ErrNotFound}
		responseBody := net.MarshalBody(res, w)
		if responseBody == nil {
			return errors.New("failed to marshal response body")
		}

		net.Respond(http.StatusNotFound, responseBody, w)
		log.Log().Info("routeGetPolicy", "msg", "not found")
		return nil
	} else {
		log.Log().Info("routeGetPolicy",
			"msg", "Failed to retrieve policy", "err", err)

		responseBody := net.MarshalBody(reqres.PolicyReadResponse{
			Err: "Internal server error"}, w,
		)
		if responseBody == nil {
			return errors.New("failed to marshal response body")
		}

		net.Respond(http.StatusInternalServerError, responseBody, w)
		log.Log().Info("routeGetPolicy", "msg", "internal server error")
		return err
	}

	responseBody := net.MarshalBody(
		reqres.PolicyReadResponse{Policy: policy}, w,
	)
	if responseBody == nil {
		return errors.New("failed to marshal response body")
	}

	net.Respond(http.StatusOK, responseBody, w)
	log.Log().Info("routeGetPolicy", "msg", "OK")

	return nil
}
