//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package secret

import (
	stdErrors "errors"
	"net/http"

	"github.com/spiffe/spike-sdk-go/api/entity/v1/reqres"
	"github.com/spiffe/spike-sdk-go/api/errors"
	"github.com/spiffe/spike-sdk-go/kv"
	"github.com/spiffe/spike/internal/net"
)

// handleGetSecretError processes errors that occur during secret retrieval
// operations and sends appropriate HTTP responses.
//
// The function distinguishes between two types of errors:
//   - kv.ErrItemNotFound: Returns HTTP 404 Not Found when the requested
//     secret does not exist
//   - Other errors: Returns HTTP 500 Internal Server Error for unexpected
//     failures during secret retrieval
//
// All errors are wrapped with additional context before being returned:
//   - Not found errors are wrapped with errors.ErrNotFound
//   - Query failures are wrapped with errors.ErrQueryFailure
//
// Parameters:
//   - err: The error that occurred during secret retrieval
//   - w: The HTTP response writer for sending error responses
//
// Returns:
//   - error: The wrapped error that was sent to the client
func handleGetSecretError(err error, w http.ResponseWriter) error {
	fName := "handleGetSecretError"

	if stdErrors.Is(err, kv.ErrItemNotFound) {
		failErr := stdErrors.Join(errors.ErrNotFound, err)
		return net.Fail(
			reqres.SecretReadNotFound, w, http.StatusNotFound, failErr, fName,
		)
	}

	failErr := stdErrors.Join(errors.ErrQueryFailure, err)
	return net.Fail(
		reqres.SecretReadInternal, w, http.StatusInternalServerError,
		failErr, fName,
	)
}

// handleGetSecretMetadataError processes errors that occur during secret
// metadata retrieval operations and sends appropriate HTTP responses.
//
// The function distinguishes between two types of errors:
//   - kv.ErrItemNotFound: Returns HTTP 404 Not Found when the requested
//     secret metadata does not exist
//   - Other errors: Returns HTTP 500 Internal Server Error for unexpected
//     failures during metadata retrieval
//
// All errors are wrapped with additional context before being returned:
//   - Not found errors are wrapped with errors.ErrNotFound
//   - Query failures are wrapped with errors.ErrQueryFailure
//
// Parameters:
//   - err: The error that occurred during secret metadata retrieval
//   - w: The HTTP response writer for sending error responses
//
// Returns:
//   - error: The wrapped error that was sent to the client
func handleGetSecretMetadataError(err error, w http.ResponseWriter) error {
	fName := "handleGetSecretMetadataError"

	if stdErrors.Is(err, kv.ErrItemNotFound) {
		failErr := stdErrors.Join(errors.ErrNotFound, err)
		return net.Fail(
			reqres.SecretMetadataNotFound, w,
			http.StatusNotFound, failErr, fName,
		)
	}

	failErr := stdErrors.Join(errors.ErrQueryFailure, err)
	return net.Fail(
		reqres.SecretMetadataInternal, w,
		http.StatusInternalServerError, failErr, fName,
	)
}
