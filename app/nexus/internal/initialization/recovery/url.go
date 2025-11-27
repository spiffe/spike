//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package recovery

import (
	"net/url"

	apiUrl "github.com/spiffe/spike-sdk-go/api/url"
	sdkErrors "github.com/spiffe/spike-sdk-go/errors"
)

// shardURL constructs the full URL for the keeper shard endpoint by joining
// the keeper API root with the shard path.
//
// This function is used during recovery operations to build the endpoint URL
// for retrieving Shamir secret shards from SPIKE Keeper instances.
//
// Parameters:
//   - keeperAPIRoot: The base URL of the keeper API
//     (e.g., "https://keeper.example.com:8443")
//
// Returns:
//   - string: The complete URL to the shard endpoint, or empty string on error
//   - *sdkErrors.SDKError: An error if URL construction fails, nil on success
//
// Example:
//
//	url, err := shardURL("https://keeper.example.com:8443")
//	// Returns: "https://keeper.example.com:8443/v1/shard", nil
func shardURL(keeperAPIRoot string) (string, *sdkErrors.SDKError) {
	u, err := url.JoinPath(keeperAPIRoot, string(apiUrl.KeeperShard))
	if err != nil {
		failErr := sdkErrors.ErrDataInvalidInput.Wrap(err)
		failErr.Msg = "failed to construct shard URL from keeper API root"
		return "", failErr
	}
	return u, nil
}
