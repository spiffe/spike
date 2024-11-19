//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package spiffe

import (
	"context"
	"errors"
	"log"

	"github.com/spiffe/go-spiffe/v2/workloadapi"

	"github.com/spiffe/spike/app/demo/internal/config"
)

// AppSpiffeSource creates and initializes a SPIFFE X.509 Source for workload
// API authentication. It retrieves the SPIFFE endpoint socket path from
// configuration and establishes a connection to obtain workload identity
// credentials.
//
// The function returns:
//   - An initialized X509Source that can be used for SPIFFE-based
//     authentication
//   - The string representation of the SPIFFE Verifiable Identity Document
//     (SVID) ID
//   - An error if the source creation or SVID retrieval fails
//
// The caller is responsible for closing the returned X509Source using
// CloseSource when it's no longer needed.
func AppSpiffeSource(ctx context.Context) (
	*workloadapi.X509Source, string, error,
) {
	socketPath := config.SpiffeEndpointSocket()

	source, err := workloadapi.NewX509Source(ctx,
		workloadapi.WithClientOptions(workloadapi.WithAddr(socketPath)))

	if err != nil {
		return nil, "", errors.Join(
			errors.New("failed to create X509Source"),
			err,
		)
	}

	svid, err := source.GetX509SVID()
	if err != nil {
		return nil, "", errors.Join(
			errors.New("unable to get X509SVID"),
			err,
		)
	}

	return source, svid.ID.String(), nil
}

// CloseSource safely closes an X509Source if it's not nil.
// Any errors encountered during closing are logged but not returned.
// It's safe to call this function with a nil source.
func CloseSource(source *workloadapi.X509Source) {
	if source == nil {
		return
	}

	if err := source.Close(); err != nil {
		log.Printf("Unable to close X509Source: %v", err)
	}
}
