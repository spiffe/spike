//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package secret

import (
	"errors"
	"net/http"

	"github.com/spiffe/spike-sdk-go/api/entity/data"
	"github.com/spiffe/spike-sdk-go/api/entity/v1/reqres"
	apiErr "github.com/spiffe/spike-sdk-go/api/errors"
	"github.com/spiffe/spike-sdk-go/kv"

	"github.com/spiffe/spike/internal/log"
	"github.com/spiffe/spike/internal/net"
)

func handleGetSecretError(err error, w http.ResponseWriter) error {
	fName := "handleGetSecretError"

	if errors.Is(err, kv.ErrItemNotFound) {
		log.Log().Info(fName, "msg", "Secret not found")

		res := reqres.SecretReadResponse{Err: data.ErrNotFound}
		responseBody := net.MarshalBody(res, w)
		if responseBody == nil {
			return apiErr.ErrMarshalFailure
		}

		net.Respond(http.StatusNotFound, responseBody, w)
		log.Log().Info("routeGetSecret", "msg", "not found")
		return nil
	}

	log.Log().Warn(fName, "msg", "Failed to retrieve secret", "err", err)

	responseBody := net.MarshalBody(reqres.SecretReadResponse{
		Err: data.ErrInternal}, w,
	)
	if responseBody == nil {
		return apiErr.ErrMarshalFailure
	}

	net.Respond(http.StatusInternalServerError, responseBody, w)
	log.Log().Error(fName, "msg", data.ErrInternal)
	return err
}

func handleGetSecretMetadataError(err error, w http.ResponseWriter) error {
	fName := "handleGetSecretMetadataError"

	if errors.Is(err, kv.ErrItemNotFound) {
		log.Log().Info(fName, "msg", "Secret not found")

		res := reqres.SecretMetadataResponse{Err: data.ErrNotFound}
		responseBody := net.MarshalBody(res, w)
		if responseBody == nil {
			return errors.New("failed to marshal response body")
		}

		net.Respond(http.StatusNotFound, responseBody, w)
		return nil
	}

	log.Log().Info(fName, "msg",
		"Failed to retrieve secret", "err", err)
	responseBody := net.MarshalBody(reqres.SecretMetadataResponse{
		Err: "Internal server error"}, w,
	)
	if responseBody == nil {
		return errors.New("failed to marshal response body")
	}

	net.Respond(http.StatusInternalServerError, responseBody, w)
	return err
}
