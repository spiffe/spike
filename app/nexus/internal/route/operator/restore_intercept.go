//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package operator

import (
	"encoding/base64"
	"github.com/cloudflare/circl/group"
	"net/http"

	"github.com/cloudflare/circl/secretsharing"
	"github.com/spiffe/spike-sdk-go/api/entity/data"
	"github.com/spiffe/spike-sdk-go/api/entity/v1/reqres"
	apiErr "github.com/spiffe/spike-sdk-go/api/errors"
	"github.com/spiffe/spike-sdk-go/spiffe"
	"github.com/spiffe/spike-sdk-go/validation"

	"github.com/spiffe/spike/internal/net"
)

func guardRestoreRequest(
	request reqres.RestoreRequest, w http.ResponseWriter, r *http.Request,
) error {
	shard := request.Shard

	spiffeid, err := spiffe.IdFromRequest(r)
	if err != nil {
		responseBody := net.MarshalBody(reqres.RestoreResponse{
			Err: data.ErrUnauthorized,
		}, w)
		net.Respond(http.StatusUnauthorized, responseBody, w)
		return apiErr.ErrUnauthorized
	}

	err = validation.ValidateSpiffeId(spiffeid.String())
	if err != nil {
		responseBody := net.MarshalBody(reqres.RestoreResponse{
			Err: data.ErrUnauthorized,
		}, w)
		net.Respond(http.StatusUnauthorized, responseBody, w)
		return apiErr.ErrUnauthorized
	}

	const expectedDecodedLength = 32 // AES-256 key size

	// Check 1: Validate that the shard is not empty
	if shard == "" {
		return apiErr.ErrInvalidInput
	}

	// Check 2: Validate that the shard is valid base64
	decodedShard, err := base64.StdEncoding.DecodeString(shard)
	if err != nil {
		return apiErr.ErrInvalidInput
	}

	// Check 3: Validate the decoded length
	if len(decodedShard) != expectedDecodedLength {
		return apiErr.ErrInvalidInput
	}

	// Check 4: The decoded share should be able to unmarshal into a
	// secretsharing.Share
	g := group.P256
	share := secretsharing.Share{
		ID:    g.NewScalar(),
		Value: g.NewScalar(),
	}
	share.ID.SetUint64(1)
	if err := share.Value.UnmarshalBinary(decodedShard); err != nil {
		return apiErr.ErrInvalidInput
	}

	return nil
}
