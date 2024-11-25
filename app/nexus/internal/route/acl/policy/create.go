//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package policy

import (
	"errors"
	"net/http"
	"time"

	state "github.com/spiffe/spike/app/nexus/internal/state/base"
	"github.com/spiffe/spike/internal/entity/data"
	"github.com/spiffe/spike/internal/entity/v1/reqres"
	"github.com/spiffe/spike/internal/log"
	"github.com/spiffe/spike/internal/net"
)

func RoutePutPolicy(
	w http.ResponseWriter, r *http.Request, audit *log.AuditEntry,
) error {
	log.Log().Info("routePutPolicy", "method", r.Method, "path", r.URL.Path,
		"query", r.URL.RawQuery)
	audit.Action = log.AuditCreate

	requestBody := net.ReadRequestBody(w, r)
	if requestBody == nil {
		return errors.New("failed to read request body")
	}

	request := net.HandleRequest[
		reqres.PolicyCreateRequest, reqres.PolicyCreateResponse](
		requestBody, w,
		reqres.PolicyCreateResponse{Err: reqres.ErrBadInput},
	)
	if request == nil {
		return errors.New("failed to parse request body")
	}

	// TODO: sanitize

	name := request.Name
	spiffeIdPattern := request.SpiffeIdPattern
	pathPattern := request.PathPattern
	permissions := request.Permissions

	policy, err := state.CreatePolicy(data.Policy{
		Id:              "",
		Name:            name,
		SpiffeIdPattern: spiffeIdPattern,
		PathPattern:     pathPattern,
		Permissions:     permissions,
		CreatedAt:       time.Time{},
		CreatedBy:       "",
	})
	if err != nil {
		log.Log().Info("routePutPolicy",
			"msg", "Failed to create policy", "err", err)

		responseBody := net.MarshalBody(reqres.PolicyCreateResponse{
			Err: "Internal server error",
		}, w)

		net.Respond(http.StatusInternalServerError, responseBody, w)
		log.Log().Error("routePutPolicy", "msg", "internal server error")

		return err
	}

	responseBody := net.MarshalBody(reqres.PolicyCreateResponse{
		Id: policy.Id,
	}, w)

	net.Respond(http.StatusOK, responseBody, w)
	log.Log().Info("routePutPolicy", "msg", "OK")

	return nil
}
