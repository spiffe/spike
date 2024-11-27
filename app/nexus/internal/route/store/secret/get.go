//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package secret

import (
	"errors"
	"net/http"

	state "github.com/spiffe/spike/app/nexus/internal/state/base"
	"github.com/spiffe/spike/internal/entity/v1/reqres"
	"github.com/spiffe/spike/internal/log"
	"github.com/spiffe/spike/internal/net"
	"github.com/spiffe/spike/pkg/store"
)

// RouteGetSecret handles requests to retrieve a secret at a specific path
// and version.
//
// This endpoint requires a valid admin JWT token for authentication. The
// function retrieves a secret based on the provided path and optional version
// number. If no version is specified, the latest version is returned.
//
// The function follows these steps:
//  1. Validates the JWT token
//  2. Validates and unmarshals the request body
//  3. Attempts to retrieve the secret
//  4. Returns the secret data or an appropriate error response
//
// Parameters:
//   - w: http.ResponseWriter to write the HTTP response
//   - r: *http.Request containing the incoming HTTP request
//   - audit: *log.AuditEntry for logging audit information
//
// Returns:
//   - error: if an error occurs during request processing.
//
// Request body format:
//
//	{
//	    "path": string,     // Path to the secret
//	    "version": int      // Optional: specific version to retrieve
//	}
//
// Response format on success (200 OK):
//
//	{
//	    "data": {          // The secret data
//	        // Secret key-value pairs
//	    }
//	}
//
// Error responses:
//   - 401 Unauthorized: Invalid or missing JWT token
//   - 400 Bad Request: Invalid request body
//   - 404 Not Found: Secret doesn't exist at specified path/version
//
// All operations are logged using structured logging.
func RouteGetSecret(
	w http.ResponseWriter, r *http.Request, audit *log.AuditEntry,
) error {
	log.Log().Info("routeGetSecret", "method", r.Method, "path", r.URL.Path,
		"query", r.URL.RawQuery)
	audit.Action = log.AuditRead

	requestBody := net.ReadRequestBody(w, r)
	if requestBody == nil {
		return errors.New("failed to read request body")
	}

	request := net.HandleRequest[
		reqres.SecretReadRequest, reqres.SecretReadResponse](
		requestBody, w,
		reqres.SecretReadResponse{Err: reqres.ErrBadInput},
	)
	if request == nil {
		return errors.New("failed to parse request body")
	}

	path := request.Path
	version := request.Version

	secret, err := state.GetSecret(path, version)
	if err != nil {
		return handleError(err, w)
	}

	responseBody := net.MarshalBody(reqres.SecretReadResponse{Data: secret}, w)

	if responseBody == nil {
		return errors.New("failed to marshal response body")
	}

	net.Respond(http.StatusOK, responseBody, w)
	log.Log().Info("routeGetSecret", "msg", "OK")
	return nil
}

// RouteGetSecretMetadata handles requests to retrieve a secret metadata at a specific path
// and version.
//
// This endpoint requires a valid admin JWT token for authentication. The
// function retrieves a secret based on the provided path and optional version
// number. If no version is specified, the latest version is returned.
//
// The function follows these steps:
//  1. Validates the JWT token
//  2. Validates and unmarshal the request body
//  3. Attempts to retrieve the secret metadata
//  4. Returns the secret metadata or an appropriate error response
//
// Parameters:
//   - w: http.ResponseWriter to write the HTTP response
//   - r: *http.Request containing the incoming HTTP request
//   - audit: *log.AuditEntry for logging audit information
//
// Returns:
//   - error: if an error occurs during request processing.
//
// Request body format:
//
//	{
//	    "path": string,     // Path to the secret
//	    "version": int      // Optional: specific version to retrieve
//	}
//
// Response format on success (200 OK):
//
// "versions": {                           // map[int]SecretMetadataVersionResponse
//
//	"version": {                          // SecretMetadataVersionResponse object
//	  "createdTime": "",                  // time.Time
//	  "version": 0,                       // int
//	  "deletedTime": null                 // *time.Time (pointer, can be null)
//	 }
//	},
//
// "metadata": {                           // SecretRawMetadataResponse object
//
//	 "currentVersion": 0,                  // int
//	 "oldestVersion": 0,                   // int
//	 "createdTime": "",                    // time.Time
//	 "updatedTime": "",                    // time.Time
//	 "maxVersions": 0                      // int
//	},
//
// "err": null                             // ErrorCode
//
// Error responses:
//   - 401 Unauthorized: Invalid or missing JWT token
//   - 400 Bad Request: Invalid request body
//   - 404 Not Found: Secret doesn't exist at specified path/version
//
// All operations are logged using structured logging.
func RouteGetSecretMetadata(
	w http.ResponseWriter, r *http.Request, audit *log.AuditEntry,
) error {
	log.Log().Info("routeGetSecretMetadata", "method", r.Method, "path", r.URL.Path,
		"query", r.URL.RawQuery)
	audit.Action = log.AuditRead

	requestBody := net.ReadRequestBody(w, r)
	if requestBody == nil {
		return errors.New("failed to read request body")
	}

	request := net.HandleRequest[
		reqres.SecretReadRequest, reqres.SecretReadResponse](
		requestBody, w,
		reqres.SecretReadResponse{Err: reqres.ErrBadInput},
	)
	if request == nil {
		return errors.New("failed to parse request body")
	}

	path := request.Path
	version := request.Version

	rawSecret, err := state.GetRawSecret(path, version)
	if err != nil {
		return handleError(err, w)
	}

	response := rawSecretResponseMapper(rawSecret)
	responseBody := net.MarshalBody(response, w)

	if responseBody == nil {
		return errors.New("failed to marshal response body")
	}

	net.Respond(http.StatusOK, responseBody, w)
	log.Log().Info("routeGetSecret", "msg", "OK")
	return nil
}

func rawSecretResponseMapper(rawSecret *store.Secret) reqres.SecretMetadataResponse {
	versions := make(map[int]reqres.SecretMetadataVersionResponse)
	for versionNum, version := range rawSecret.Versions {
		versions[versionNum] = reqres.SecretMetadataVersionResponse{
			CreatedTime: version.CreatedTime,
			Version:     version.Version,
			DeletedTime: version.DeletedTime,
		}
	}

	metadata := reqres.SecretRawMetadataResponse{
		CurrentVersion: rawSecret.Metadata.CurrentVersion,
		OldestVersion:  rawSecret.Metadata.OldestVersion,
		CreatedTime:    rawSecret.Metadata.CreatedTime,
		UpdatedTime:    rawSecret.Metadata.UpdatedTime,
		MaxVersions:    rawSecret.Metadata.MaxVersions,
	}

	return reqres.SecretMetadataResponse{
		Versions: versions,
		Metadata: metadata,
	}
}

func handleError(err error, w http.ResponseWriter) error {
	if errors.Is(err, store.ErrSecretNotFound) {
		log.Log().Info("routeGetSecret", "msg", "Secret not found")

		res := reqres.SecretReadResponse{Err: reqres.ErrNotFound}
		responseBody := net.MarshalBody(res, w)
		if responseBody == nil {
			return errors.New("failed to marshal response body")
		}
		net.Respond(http.StatusNotFound, responseBody, w)
		return nil
	}

	log.Log().Info("routeGetSecret", "msg", "Failed to retrieve secret", "err", err)
	responseBody := net.MarshalBody(reqres.SecretReadResponse{
		Err: "Internal server error"}, w,
	)
	if responseBody == nil {
		return errors.New("failed to marshal response body")
	}
	net.Respond(http.StatusInternalServerError, responseBody, w)
	return err
}
