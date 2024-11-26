//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package spiffe

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"

	"github.com/spiffe/go-spiffe/v2/spiffeid"
	"github.com/spiffe/go-spiffe/v2/svid/x509svid"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
)

// EndpointSocket returns the UNIX domain socket address for the SPIFFE
// Workload API endpoint.
//
// The function first checks for the SPIFFE_ENDPOINT_SOCKET environment
// variable. If set, it returns that value. Otherwise, it returns a default
// development
//
//	socket path:
//
// "unix:///tmp/spire-agent/public/api.sock"
//
// For production deployments, especially in Kubernetes environments, it's
// recommended to set SPIFFE_ENDPOINT_SOCKET to a more restricted socket path,
// such as: "unix:///run/spire/agent/sockets/spire.sock"
//
// Default socket paths by environment:
//   - Development (Linux): unix:///tmp/spire-agent/public/api.sock
//   - Kubernetes: unix:///run/spire/agent/sockets/spire.sock
//
// Returns:
//   - string: The UNIX domain socket address for the SPIFFE Workload API
//     endpoint
//
// Environment Variables:
//   - SPIFFE_ENDPOINT_SOCKET: Override the default socket path
func EndpointSocket() string {
	p := os.Getenv("SPIFFE_ENDPOINT_SOCKET")
	if p != "" {
		return p
	}

	return "unix:///tmp/spire-agent/public/api.sock"
}

// AppSpiffeSource creates and initializes a new X509Source for SPIFFE
// authentication.
//
// The function establishes a connection to the SPIRE Agent through a Unix
// domain socket and retrieves the X509-SVID (SPIFFE Verifiable Identity
// Document) for the current workload. This is typically used during application
// startup to set up SPIFFE-based authentication.
//
// Parameters:
//   - ctx: Context for controlling the source creation lifecycle
//
// Returns:
//   - *workloadapi.X509Source: The initialized X509 source for SPIFFE
//     authentication
//   - string: The SPIFFE ID string associated with the workload's X509-SVID
//
// The function will call return an error if it encounters errors during:
//   - X509Source creation
//   - X509-SVID retrieval
func AppSpiffeSource(ctx context.Context) (
	*workloadapi.X509Source, string, error,
) {
	socketPath := EndpointSocket()

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

// IdFromRequest extracts the SPIFFE ID from the TLS peer certificate of
// an HTTP request.
// It checks if the incoming request has a valid TLS connection and at least one
// peer certificate.
// The first certificate in the chain is used to extract the SPIFFE ID.
//
// Params:
//
//	r *http.Request - The HTTP request from which the SPIFFE ID is to be
//	extracted.
//
// Returns:
//
//	 *spiffeid.ID - The SPIFFE ID extracted from the first peer certificate,
//	 or nil if extraction fails.
//	 error - An error object indicating the failure reason. Possible errors
//	include the absence of peer certificates or a failure in extracting the
//	SPIFFE ID from the certificate.
//
// Note:
//
//	This function assumes that the request is already over a secured TLS
//	connection and will fail if the TLS connection state is not available or
//	the peer certificates are missing.
func IdFromRequest(r *http.Request) (*spiffeid.ID, error) {
	tlsConnectionState := r.TLS
	if len(tlsConnectionState.PeerCertificates) == 0 {
		return nil, errors.New("no peer certs")
	}

	id, err := x509svid.IDFromCert(tlsConnectionState.PeerCertificates[0])
	if err != nil {
		return nil, errors.Join(
			err,
			errors.New("problem extracting svid"),
		)
	}

	return &id, nil
}

// CloseSource safely closes an X509Source.
//
// This function should be called when the X509Source is no longer needed,
// typically during application shutdown or cleanup. It handles nil sources
// gracefully and logs any errors that occur during closure without failing.
//
// Parameters:
//   - source: The X509Source to close, may be nil
//
// If an error occurs during closure, it will be logged but will not cause the
// function to panic or return an error.
func CloseSource(source *workloadapi.X509Source) {
	if source == nil {
		return
	}

	if err := source.Close(); err != nil {
		log.Printf("Unable to close X509Source: %v", err)
	}
}
