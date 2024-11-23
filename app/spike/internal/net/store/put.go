//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package store

import (
	"encoding/json"
	"errors"

	"github.com/spiffe/go-spiffe/v2/workloadapi"

	"github.com/spiffe/spike/app/spike/internal/net/api"
	"github.com/spiffe/spike/internal/auth"
	"github.com/spiffe/spike/internal/entity/v1/reqres"
	"github.com/spiffe/spike/internal/net"
)

// PutSecret creates or updates a secret at the specified path with the given
// values using mTLS authentication.
//
// Parameters:
//   - source: X509Source for mTLS client authentication
//   - path: Path where the secret should be stored
//   - values: Map of key-value pairs representing the secret data
//
// Returns:
//   - error: nil on success, unauthorized error if not logged in, or
//     wrapped error on request/parsing failure
//
// Example:
//
//	err := PutSecret(x509Source, "secret/path", map[string]string{"key": "value"})
func PutSecret(source *workloadapi.X509Source,
	path string, values map[string]string) error {

	r := reqres.SecretPutRequest{
		Path:   path,
		Values: values,
	}

	mr, err := json.Marshal(r)
	if err != nil {
		return errors.Join(
			errors.New("putSecret: I am having problem generating the payload"),
			err,
		)
	}

	client, err := net.CreateMtlsClient(source, auth.IsNexus)
	if err != nil {
		return err
	}

	_, err = net.Post(client, api.UrlSecretPut(), mr)
	if errors.Is(err, net.ErrUnauthorized) {
		return errors.New(`unauthorized. Please login first with 'spike login'`)
	}

	return err
}
