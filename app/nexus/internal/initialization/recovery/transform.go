//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package recovery

import (
	"bytes"
	"encoding/json"
	"net/url"
	"strings"

	"github.com/spiffe/spike-sdk-go/api/entity/v1/reqres"
	"github.com/spiffe/spike-sdk-go/crypto"
	sdkErrors "github.com/spiffe/spike-sdk-go/errors"
	"github.com/spiffe/spike-sdk-go/security/mem"
)

// keeperURL joins a keeper API root with a keeper API path after validating
// the root.
//
// The root must be an absolute https URL with a host, because every keeper
// exchange runs over mTLS. An empty root, a relative path, a plain http
// root, or a root that cannot be parsed is rejected here, so that a
// misconfigured keeper entry never turns into an outbound request to an
// unintended destination.
//
// Parameters:
//   - keeperAPIRoot: The keeper's base URL from the configuration.
//   - apiPath: The keeper API path to append to the root.
//
// Returns:
//   - string: The joined URL on success, empty on failure.
//   - *sdkErrors.SDKError: ErrDataInvalidInput when the root is empty or is
//     not an absolute https URL, or ErrAPIBadRequest when the join fails.
func keeperURL(keeperAPIRoot, apiPath string) (string, *sdkErrors.SDKError) {
	if strings.TrimSpace(keeperAPIRoot) == "" {
		failErr := sdkErrors.ErrDataInvalidInput.Clone()
		failErr.Msg = "keeper API root is empty"
		return "", failErr
	}

	parsed, parseErr := url.Parse(keeperAPIRoot)
	if parseErr != nil {
		failErr := sdkErrors.ErrDataInvalidInput.Wrap(parseErr)
		failErr.Msg = "keeper API root is not a valid URL"
		return "", failErr
	}
	if parsed.Scheme != "https" || parsed.Host == "" {
		failErr := sdkErrors.ErrDataInvalidInput.Clone()
		failErr.Msg = "keeper API root must be an absolute https URL: " +
			keeperAPIRoot
		return "", failErr
	}

	joined, joinErr := url.JoinPath(keeperAPIRoot, apiPath)
	if joinErr != nil {
		failErr := sdkErrors.ErrAPIBadRequest.Wrap(joinErr)
		failErr.Msg = "failed to join path"
		return "", failErr
	}

	return joined, nil
}

// unmarshalShardResponse parses a keeper's shard response and rejects any
// body that does not carry a shard.
//
// The decoder refuses unknown fields and trailing data, a JSON null or a
// body without a "shard" member is rejected, and a keeper-reported error
// code is surfaced instead of being silently accepted as an empty shard.
//
// Parameters:
//   - data: The raw JSON response body received from the keeper.
//
// Returns:
//   - *reqres.ShardGetResponse: The parsed response with a non-nil Shard.
//   - *sdkErrors.SDKError: ErrDataUnmarshalFailure when the body is not a
//     well-formed shard response, ErrAPIInternal when the keeper reported an
//     error code, or ErrShamirEmptyShard when no shard is present.
func unmarshalShardResponse(data []byte) (
	*reqres.ShardGetResponse, *sdkErrors.SDKError,
) {
	var res reqres.ShardGetResponse

	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if decodeErr := dec.Decode(&res); decodeErr != nil {
		failErr := sdkErrors.ErrDataUnmarshalFailure.Wrap(decodeErr)
		failErr.Msg = "failed to unmarshal response"
		return nil, failErr
	}
	if dec.More() {
		failErr := sdkErrors.ErrDataUnmarshalFailure.Clone()
		failErr.Msg = "unexpected trailing data in shard response"
		return nil, failErr
	}

	if res.Err != "" {
		failErr := sdkErrors.ErrAPIInternal.Clone()
		failErr.Msg = "keeper reported an error: " + string(res.Err)
		return nil, failErr
	}

	if res.Shard == nil {
		failErr := sdkErrors.ErrShamirEmptyShard.Clone()
		failErr.Msg = "shard response carries no shard"
		return nil, failErr
	}

	return &res, nil
}

// resetShards zeroes every shard in the map and removes all entries, so the
// next recovery round starts from an empty, consistent set.
//
// Parameters:
//   - shards: The shards collected so far, keyed by keeper ID.
func resetShards(shards map[string]*[crypto.AES256KeySize]byte) {
	for id, shard := range shards {
		mem.ClearRawBytes(shard)
		delete(shards, id)
	}
}
