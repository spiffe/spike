//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\ SPDX-License-Identifier: Apache-2.0

package cipher

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"net/http"

	"github.com/spiffe/go-spiffe/v2/workloadapi"
	apiUrl "github.com/spiffe/spike-sdk-go/api/url"
	"github.com/spiffe/spike/app/spike/internal/env"
)

// streamToCipherEndpoint posts a binary stream to the given cipher API path
// using SPIFFE-based mTLS. It returns the HTTP response for the caller to copy
// the body to the desired writer.
func streamToCipherEndpoint(
	source *workloadapi.X509Source,
	path apiUrl.APIURL,
	in io.Reader,
) (*http.Response, error) {
	endpoint := env.NexusBaseURL() + string(path)

	svid, err := source.GetX509SVID()
	if err != nil {
		return nil, fmt.Errorf("failed to get X509 SVID: %w", err)
	}

	tlsCert := tls.Certificate{Certificate: [][]byte{}, PrivateKey: svid.PrivateKey}
	for _, c := range svid.Certificates {
		tlsCert.Certificate = append(tlsCert.Certificate, c.Raw)
	}

	roots := x509.NewCertPool()
	bundle, err := source.GetX509BundleForTrustDomain(svid.ID.TrustDomain())
	if err != nil {
		return nil, fmt.Errorf("failed to get X509 bundle: %w", err)
	}
	for _, r := range bundle.X509Authorities() {
		roots.AddCert(r)
	}

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			Certificates: []tls.Certificate{tlsCert},
			RootCAs:      roots,
			MinVersion:   tls.VersionTLS12,
		},
	}
	client := &http.Client{Transport: transport}

	req, err := http.NewRequest("POST", endpoint, in)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/octet-stream")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call endpoint: %w", err)
	}
	return resp, nil
}
