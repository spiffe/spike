//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package cipher

import (
	"crypto/cipher"
	"net/http"

	"github.com/spiffe/spike-sdk-go/api/entity/v1/reqres"
	sdkErrors "github.com/spiffe/spike-sdk-go/errors"
	"github.com/spiffe/spike-sdk-go/net"
	"github.com/spiffe/spike-sdk-go/predicate"

	state "github.com/spiffe/spike/app/nexus/internal/state/base"
)

// handleStreamingDecrypt processes a complete streaming mode decryption
// request, including reading, validating, decrypting, and responding.
//
// The cipher is retrieved only after SPIFFE ID validation passes, following
// the principle of least privilege. Full request validation (including
// request fields) happens after the request is constructed.
//
// Parameters:
//   - w: The HTTP response writer
//   - r: The HTTP request
//   - getCipher: Function to retrieve the cipher after authentication
//
// Returns:
//   - *sdkErrors.SDKError: An error if any step fails
func handleStreamingDecrypt(
	w http.ResponseWriter, r *http.Request,
	getCipher func() (cipher.AEAD, *sdkErrors.SDKError),
) *sdkErrors.SDKError {
	// NOTE: since we are dealing with streaming data, we cannot directly use
	// the request parameter validation patterns that we employ in the JSON/REST
	// payloads. We need to read the entire stream and generate a request
	// entity accordingly.

	if authErr := net.AuthorizeAndRespondOnFail(
		reqres.CipherDecryptResponse{}.Unauthorized(),
		predicate.AllowSPIFFEIDForCipherDecrypt,
		state.CheckPolicyAccess,
		w, r,
	); authErr != nil {
		return authErr
	}

	// Get cipher only after SPIFFE ID validation passes
	c, cipherErr := getCipher()
	if cipherErr != nil {
		return cipherErr
	}

	// Read request data (now that we have cipher for nonce size)
	version, nonce, ciphertext, readErr := readStreamingDecryptRequestData(
		w, r, c,
	)
	if readErr != nil {
		return readErr
	}

	// Construct request object for guard validation
	request := reqres.CipherDecryptRequest{
		Version:    version,
		Nonce:      nonce,
		Ciphertext: ciphertext,
	}

	// Full guard validation (auth and request fields)
	guardErr := guardCipherDecryptRequest(request, w, r)
	if guardErr != nil {
		return guardErr
	}

	plaintext, decryptErr := net.DecryptDataStreaming(nonce, ciphertext, c, w)
	if decryptErr != nil {
		return decryptErr
	}

	return respondStreamingDecrypt(plaintext, w)
}

// handleJSONDecrypt processes a complete JSON mode decryption request,
// including reading, validating, decrypting, and responding.
//
// The cipher is retrieved only after SPIFFE ID validation passes, following
// the principle of least privilege. Full request validation (including
// request fields) happens after the request is parsed.
//
// Parameters:
//   - w: The HTTP response writer
//   - r: The HTTP request
//   - getCipher: Function to retrieve the cipher after authentication
//
// Returns:
//   - *sdkErrors.SDKError: An error if any step fails
func handleJSONDecrypt(
	w http.ResponseWriter, r *http.Request,
	getCipher func() (cipher.AEAD, *sdkErrors.SDKError),
) *sdkErrors.SDKError {
	// Extract and validate SPIFFE ID before accessing cipher
	if authErr := net.AuthorizeAndRespondOnFail(
		reqres.CipherDecryptResponse{}.Unauthorized(),
		predicate.AllowSPIFFEIDForCipherDecrypt,
		state.CheckPolicyAccess,
		w, r,
	); authErr != nil {
		return authErr
	}

	// Parse request (doesn't need cipher)
	request, readErr := readJSONDecryptRequestWithoutGuard(w, r)
	if readErr != nil {
		return readErr
	}

	// Full guard validation (auth and request fields)
	guardErr := guardCipherDecryptRequest(*request, w, r)
	if guardErr != nil {
		return guardErr
	}

	// Get cipher only after auth passes
	c, cipherErr := getCipher()
	if cipherErr != nil {
		return cipherErr
	}

	plaintext, decryptErr := net.DecryptDataJSON(
		request.Nonce, request.Ciphertext, c, w,
	)
	if decryptErr != nil {
		return decryptErr
	}

	return respondJSONDecrypt(plaintext, w)
}

// handleStreamingEncrypt processes a complete streaming mode encryption
// request, including reading, nonce generation, encrypting, and responding.
//
// The cipher is retrieved only after SPIFFE ID validation passes, following
// the principle of least privilege. Full request validation (including
// request fields) happens after the request is constructed.
//
// Parameters:
//   - w: The HTTP response writer
//   - r: The HTTP request
//   - getCipher: Function to retrieve the cipher after authentication
//
// Returns:
//   - *sdkErrors.SDKError: An error if any step fails
func handleStreamingEncrypt(
	w http.ResponseWriter, r *http.Request,
	getCipher func() (cipher.AEAD, *sdkErrors.SDKError),
) *sdkErrors.SDKError {
	req, err := readAndGuardRequest(
		readStreamingEncryptRequestWithoutGuard,
		guardCipherEncryptRequest,
		w, r,
	)
	if err != nil {
		return err
	}

	nonce, ciphertext, encryptErr := getCipherAndEncrypt(getCipher, net.EncryptDataStreaming, req.Plaintext, w)
	if encryptErr != nil {
		return encryptErr
	}

	return respondStreamingEncrypt(nonce, ciphertext, w)
}

type Handler[T any] func(w http.ResponseWriter, r *http.Request) (*T, *sdkErrors.SDKError)
type HandlerWithEntity[T any] func(req T, w http.ResponseWriter, r *http.Request) *sdkErrors.SDKError

type Encryptor func(plaintext []byte, c cipher.AEAD, w http.ResponseWriter) ([]byte, []byte, *sdkErrors.SDKError)

// readAndGuardRequest reads and parses a request, then validates it using the
// provided guard function. This is similar to net.ReadParseAndGuard but accepts
// a custom reader function for streaming mode support.
//
// Parameters:
//   - readRequest: Function to read and parse the request body
//   - guard: Function to validate the parsed request (handles auth and fields)
//   - w: The HTTP response writer
//   - r: The HTTP request
//
// Returns:
//   - *T: The parsed and validated request
//   - *sdkErrors.SDKError: An error if reading or validation fails
func readAndGuardRequest[T any](
	readRequest Handler[T],
	guard HandlerWithEntity[T],
	w http.ResponseWriter, r *http.Request,
) (*T, *sdkErrors.SDKError) {
	request, readErr := readRequest(w, r)
	if readErr != nil {
		return nil, readErr
	}

	if guardErr := guard(*request, w, r); guardErr != nil {
		return nil, guardErr
	}

	return request, nil
}

// getCipherAndEncrypt retrieves the cipher and encrypts the provided data.
// This combines cipher acquisition and encryption into a single operation.
//
// Parameters:
//   - getCipher: Function to retrieve the AEAD cipher
//   - encryptData: The encryption function to use
//   - plaintext: The data to encrypt
//   - w: The HTTP response writer for error responses
//
// Returns:
//   - []byte: The generated nonce
//   - []byte: The encrypted ciphertext
//   - *sdkErrors.SDKError: An error if cipher retrieval or encryption fails
func getCipherAndEncrypt(
	getCipher func() (cipher.AEAD, *sdkErrors.SDKError),
	encryptData Encryptor,
	plaintext []byte,
	w http.ResponseWriter,
) ([]byte, []byte, *sdkErrors.SDKError) {
	c, cipherErr := getCipher()
	if cipherErr != nil {
		return nil, nil, cipherErr
	}

	return encryptData(plaintext, c, w)
}

// handleJSONEncrypt processes a complete JSON mode encryption request,
// including reading, nonce generation, encrypting, and responding.
//
// The cipher is retrieved only after SPIFFE ID validation passes, following
// the principle of least privilege. Full request validation (including
// request fields) happens after the request is parsed.
//
// Parameters:
//   - w: The HTTP response writer
//   - r: The HTTP request
//   - getCipher: Function to retrieve the cipher after authentication
//
// Returns:
//   - *sdkErrors.SDKError: An error if any step fails
func handleJSONEncrypt(
	w http.ResponseWriter, r *http.Request,
	getCipher func() (cipher.AEAD, *sdkErrors.SDKError),
) *sdkErrors.SDKError {
	req, err := readAndGuardRequest(
		readJSONEncryptRequestWithoutGuard,
		guardCipherEncryptRequest,
		w, r,
	)
	if err != nil {
		return err
	}

	nonce, ciphertext, encryptErr := getCipherAndEncrypt(getCipher, net.EncryptDataJSON, req.Plaintext, w)
	if encryptErr != nil {
		return encryptErr
	}

	return respondJSONEncrypt(nonce, ciphertext, w)
}
