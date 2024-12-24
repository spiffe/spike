//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package secret

import (
	"github.com/spiffe/spike-sdk-go/api/entity/data"
	"github.com/spiffe/spike-sdk-go/api/entity/v1/reqres"
	"github.com/spiffe/spike-sdk-go/spiffe"
	"github.com/spiffe/spike-sdk-go/validation"
	"github.com/spiffe/spike/internal/net"
	"net/http"
)

func guardPutSecretMetadataRequest(
	request reqres.SecretPutRequest, w http.ResponseWriter, r *http.Request,
) error {

	spiffeid, err := spiffe.IdFromRequest(r)
	if err != nil {
		responseBody := net.MarshalBody(reqres.SecretPutResponse{
			Err: data.ErrUnauthorized,
		}, w)
		net.Respond(http.StatusUnauthorized, responseBody, w)
		return err
	}
	err = validation.ValidateSpiffeId(spiffeid.String())
	if err != nil {
		responseBody := net.MarshalBody(reqres.SecretPutResponse{
			Err: data.ErrUnauthorized,
		}, w)
		net.Respond(http.StatusUnauthorized, responseBody, w)
	}

	return nil
}
