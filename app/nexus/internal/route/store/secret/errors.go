//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package secret

import (
	"errors"
	"net/http"

	"github.com/spiffe/spike-sdk-go/api/entity/data"
	"github.com/spiffe/spike-sdk-go/api/entity/v1/reqres"

	"github.com/spiffe/spike/internal/log"
	"github.com/spiffe/spike/internal/net"
	"github.com/spiffe/spike/pkg/store"
)

func handleGetSecretError(err error, w http.ResponseWriter) error {
	// TODO: maybe reuse this in getSecret too -- currently only getsecretmeta uses it.

	fName := "handleGetSecretError"

	if errors.Is(err, store.ErrSecretNotFound) {
		log.Log().Info(fName, "msg", "Secret not found")

		res := reqres.SecretReadResponse{Err: data.ErrNotFound}
		responseBody := net.MarshalBody(res, w)
		if responseBody == nil {
			return errors.New("failed to marshal response body")
		}

		net.Respond(http.StatusNotFound, responseBody, w)
		return nil
	}

	log.Log().Info(fName, "msg",
		"Failed to retrieve secret", "err", err)
	responseBody := net.MarshalBody(reqres.SecretReadResponse{
		Err: "Internal server error"}, w,
	)
	if responseBody == nil {
		return errors.New("failed to marshal response body")
	}

	net.Respond(http.StatusInternalServerError, responseBody, w)
	return err
}
