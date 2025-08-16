//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package url

import (
	"net/url"
	"os"

	apiUrl "github.com/spiffe/spike-sdk-go/api/url"
	"github.com/spiffe/spike-sdk-go/log"
)

// KeeperEndpoint constructs the full API endpoint URL for keeper contribution
// requests. It joins the provided keeper API root URL with the
// KeeperContribute path segment to create a complete endpoint URL for
// submitting secret shares to keepers. The function will terminate the program
// with exit code 1 if URL path joining fails.
func KeeperEndpoint(keeperAPIRoot string) string {
	const fName = "keeperEndpoint"

	u, err := url.JoinPath(
		keeperAPIRoot, string(apiUrl.KeeperContribute),
	)
	if err != nil {
		log.Log().Warn(
			fName, "message", "Failed to join path", "url", keeperAPIRoot,
		)
		os.Exit(1)
	}
	return u
}
