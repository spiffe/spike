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

func CloseSource(source *workloadapi.X509Source) {
	if source == nil {
		return
	}

	if err := source.Close(); err != nil {
		log.Printf("Unable to close X509Source: %v", err)
	}
}
