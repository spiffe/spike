//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

// Package auth provides authentication utilities for SPIFFE-based operations
// in SPIKE. It offers functions for extracting and validating SPIFFE IDs from
// HTTP requests, enabling secure peer authentication in the SPIKE ecosystem.
package auth

import (
	"net/http"

	"github.com/spiffe/go-spiffe/v2/spiffeid"
	apiErr "github.com/spiffe/spike-sdk-go/api/errors"
	"github.com/spiffe/spike-sdk-go/spiffe"
	"github.com/spiffe/spike-sdk-go/validation"

	"github.com/spiffe/spike/internal/net"
)

// ExtractPeerSPIFFEID extracts and validates the peer SPIFFE ID from an HTTP
// request. If the SPIFFE ID cannot be extracted or is nil, it writes an
// unauthorized response using the provided error response object and returns
// an error.
//
// This function is generic and can be used with any response type that needs
// to be returned in case of authentication failure.
//
// Parameters:
//   - r *http.Request: The HTTP request containing peer SPIFFE ID
//   - w http.ResponseWriter: Response writer for error responses
//   - errorResponse T: The error response object to marshal and send if
//     validation fails
//
// Returns:
//   - *spiffeid.ID: The extracted SPIFFE ID if successful
//   - error: apiErr.ErrUnauthorized if extraction fails or ID is nil,
//     nil otherwise
//
// Example usage:
//
//	peerID, err := auth.ExtractPeerSPIFFEID(
//	    r, w,
//	    reqres.ShardGetResponse{Err: data.ErrUnauthorized},
//	)
//	if err != nil {
//	    return err
//	}
func ExtractPeerSPIFFEID[T any](
	r *http.Request,
	w http.ResponseWriter,
	errorResponse T,
) (*spiffeid.ID, error) {
	peerSPIFFEID, err := spiffe.IDFromRequest(r)
	if err != nil {
		responseBody, err := net.MarshalBodyAndRespondOnMarshalFail(
			errorResponse, w,
		)
		if err == nil {
			net.Respond(http.StatusUnauthorized, responseBody, w)
		}
		return nil, apiErr.ErrUnauthorized
	}

	if peerSPIFFEID == nil {
		responseBody, err := net.MarshalBodyAndRespondOnMarshalFail(
			errorResponse, w,
		)
		if err == nil {
			net.Respond(http.StatusUnauthorized, responseBody, w)
		}
		return nil, apiErr.ErrUnauthorized
	}

	err = validation.ValidateSPIFFEID(peerSPIFFEID.String())
	if err != nil {
		responseBody, err := net.MarshalBodyAndRespondOnMarshalFail(
			errorResponse, w,
		)
		if err == nil {
			net.Respond(http.StatusUnauthorized, responseBody, w)
		}
		return nil, apiErr.ErrUnauthorized
	}

	return peerSPIFFEID, nil
}
