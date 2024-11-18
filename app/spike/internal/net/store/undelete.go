//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package store

import (
	"encoding/json"
	"errors"
	net2 "github.com/spiffe/spike/app/spike/internal/net"
	"github.com/spiffe/spike/internal/auth"
	"strconv"

	"github.com/spiffe/go-spiffe/v2/workloadapi"

	"github.com/spiffe/spike/internal/entity/v1/reqres"
	"github.com/spiffe/spike/internal/net"
)

// UndeleteSecret undeletes a secret from SPIKE Nexus.
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

	_, err = net.Post(client, net2.UrlSecretUndelete(), mr)
	if errors.Is(err, net.ErrUnauthorized) {
		return errors.New(`unauthorized. Please login first with 'spike login'`)
	}

	return err
}
