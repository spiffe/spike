package secret

import (
	"fmt"
	"net/http"

	"github.com/spiffe/spike-sdk-go/api/entity/data"
	"github.com/spiffe/spike-sdk-go/api/entity/v1/reqres"
	apiErr "github.com/spiffe/spike-sdk-go/api/errors"
	"github.com/spiffe/spike-sdk-go/log"

	"github.com/spiffe/spike/app/nexus/internal/state/persist"
	journal "github.com/spiffe/spike/internal/log"
	"github.com/spiffe/spike/internal/net"
)

// RouteDecrypt handles HTTP requests to decrypt ciphertext data using the
// system's cipher. This endpoint provides decryption-as-a-service functionality
// without persisting any data.
//
// The function expects a JSON request body containing:
//   - version: protocol version byte (must be '1')
//   - nonce: the nonce used during encryption
//   - ciphertext: encrypted data to decrypt
//   - algorithm: (optional) decryption algorithm to use
//
// On success, it returns a JSON response containing:
//   - plaintext: the decrypted data
//   - err: error code (ErrSuccess on success)
//
// The decryption process:
//  1. Validates and parses the incoming request
//  2. Checks access permissions via guardDecryptSecretRequest
//  3. Retrieves the system cipher from the backend
//  4. Validates the protocol version
//  5. Decrypts the ciphertext using authenticated decryption (AEAD)
//  6. Returns the decrypted plaintext
//
// Access control is enforced through guardDecryptSecretRequest, which should
// verify the caller has appropriate permissions to use the decryption service.
//
// Errors:
//   - Returns ErrReadFailure if request body cannot be read
//   - Returns ErrParseFailure if request cannot be parsed as SecretDecryptRequest
//   - Returns ErrBadInput if version is not supported
//   - Returns ErrInternal if cipher is unavailable or decryption fails
//   - Returns ErrMarshalFailure if response cannot be marshaled
//   - May return errors from guardDecryptSecretRequest for permission failures
func RouteDecrypt(
	w http.ResponseWriter, r *http.Request, audit *journal.AuditEntry,
) error {
	const fName = "routeDecrypt"
	journal.AuditRequest(fName, r, audit, journal.AuditCreate)

	requestBody := net.ReadRequestBody(w, r)
	if requestBody == nil {
		return apiErr.ErrReadFailure
	}

	request := net.HandleRequest[
		reqres.SecretDecryptRequest, reqres.SecretDecryptResponse](
		requestBody, w,
		reqres.SecretDecryptResponse{Err: data.ErrBadInput},
	)
	if request == nil {
		return apiErr.ErrParseFailure
	}

	err := guardDecryptSecretRequest(*request, w, r)
	if err != nil {
		return err
	}

	// Validate version
	if request.Version != byte('1') {
		responseBody := net.MarshalBody(reqres.SecretDecryptResponse{
			Err: data.ErrBadInput,
		}, w)
		if responseBody == nil {
			return apiErr.ErrMarshalFailure
		}
		net.Respond(http.StatusBadRequest, responseBody, w)
		return fmt.Errorf("unsupported version: %v", request.Version)
	}

	// Get cipher from the backend
	c := persist.Backend().GetCipher()
	if c == nil {
		responseBody := net.MarshalBody(reqres.SecretDecryptResponse{
			Err: data.ErrInternal,
		}, w)
		if responseBody == nil {
			return apiErr.ErrMarshalFailure
		}
		net.Respond(http.StatusInternalServerError, responseBody, w)
		return fmt.Errorf("cipher not available")
	}

	// Validate nonce size
	if len(request.Nonce) != c.NonceSize() {
		responseBody := net.MarshalBody(reqres.SecretDecryptResponse{
			Err: data.ErrBadInput,
		}, w)
		if responseBody == nil {
			return apiErr.ErrMarshalFailure
		}
		net.Respond(http.StatusBadRequest, responseBody, w)
		return fmt.Errorf("invalid nonce size: expected %d, got %d",
			c.NonceSize(), len(request.Nonce))
	}

	// Decrypt the ciphertext
	log.Log().Info(fName, "message",
		fmt.Sprintf("Decrypt %d %d", len(request.Nonce), len(request.Ciphertext)),
	)

	plaintext, err := c.Open(nil, request.Nonce, request.Ciphertext, nil)
	if err != nil {
		log.Log().Info(fName, "message", fmt.Errorf("failed to decrypt %w", err))
		responseBody := net.MarshalBody(reqres.SecretDecryptResponse{
			Err: data.ErrInternal,
		}, w)
		if responseBody == nil {
			return apiErr.ErrMarshalFailure
		}
		net.Respond(http.StatusBadRequest, responseBody, w)
		return fmt.Errorf("decryption failed: %w", err)
	}

	// Prepare response
	responseBody := net.MarshalBody(reqres.SecretDecryptResponse{
		Plaintext: plaintext,
		Err:       data.ErrSuccess,
	}, w)
	if responseBody == nil {
		return apiErr.ErrMarshalFailure
	}

	net.Respond(http.StatusOK, responseBody, w)
	log.Log().Info(fName, "message", data.ErrSuccess)
	return nil
}
