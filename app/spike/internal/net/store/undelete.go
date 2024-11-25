//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package store

import (
	"encoding/json"
	"errors"

	"strconv"

	"github.com/spiffe/go-spiffe/v2/workloadapi"

	"github.com/spiffe/spike/app/spike/internal/net/api"
	"github.com/spiffe/spike/internal/auth"
	"github.com/spiffe/spike/internal/entity/v1/reqres"
	"github.com/spiffe/spike/internal/net"
)

// UndeleteSecret restores previously deleted versions of a secret at the
// specified path using mTLS authentication.
//
// Parameters:
//   - source: X509Source for mTLS client authentication
//   - path: Path to the secret to restore
//   - versions: String array of version numbers to restore. Empty array
//     attempts no restoration
//
// Returns:
//   - error: nil on success, unauthorized error if not logged in, or
//     wrapped error on request/parsing failure
//
// Example:
//
//	err := UndeleteSecret(x509Source, "secret/path", []string{"1", "2"})
func UndeleteSecret(source *workloadapi.X509Source,
	path string, versions []string) error {
	var vv []int
	if len(versions) == 0 {
		vv = []int{}
	}

	for _, version := range versions {
		v, e := strconv.Atoi(version)
		if e != nil {
			continue
		}
		vv = append(vv, v)
	}

	r := reqres.SecretUndeleteRequest{
		Path:     path,
		Versions: vv,
	}

	mr, err := json.Marshal(r)
	if err != nil {
		return errors.Join(
			errors.New(
				"undeleteSecret: I am having problem generating the payload",
			),
			err,
		)
	}

	client, err := net.CreateMtlsClient(source, auth.IsNexus)
	if err != nil {
		return err
	}

	_, err = net.Post(client, api.UrlSecretUndelete(), mr)
	return err
}
