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

	// Extract and validate SPIFFE ID before accessing cipher
	peerSPIFFEID, err := net.ExtractPeerSPIFFEIDAndRespondOnFail(
		r, w, reqres.CipherDecryptResponse{
			Err: sdkErrors.ErrAccessUnauthorized.Code,
		})
	if err != nil {
		return err
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
	guardErr := guardDecryptCipherRequest(request, peerSPIFFEID, w, r)
	if guardErr != nil {
		return guardErr
	}

	plaintext, decryptErr := decryptDataStreaming(nonce, ciphertext, c, w)
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
	peerSPIFFEID, err := extractAndValidateSPIFFEID(w, r)
	if err != nil {
		return err
	}

	// Parse request (doesn't need cipher)
	request, readErr := readJSONDecryptRequestWithoutGuard(w, r)
	if readErr != nil {
		return readErr
	}

	// Full guard validation (auth and request fields)
	guardErr := guardDecryptCipherRequest(*request, peerSPIFFEID, w, r)
	if guardErr != nil {
		return guardErr
	}

	// Get cipher only after auth passes
	c, cipherErr := getCipher()
	if cipherErr != nil {
		return cipherErr
	}

	plaintext, decryptErr := decryptDataJSON(
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
	// Extract and validate SPIFFE ID before accessing cipher
	peerSPIFFEID, err := extractAndValidateSPIFFEID(w, r)
	if err != nil {
		return err
	}

	// Read plaintext (doesn't need cipher)
	plaintext, readErr := readStreamingEncryptRequestWithoutGuard(w, r)
	if readErr != nil {
		return readErr
	}

	// Construct request object for guard validation
	request := reqres.CipherEncryptRequest{
		Plaintext: plaintext,
	}

	// Full guard validation (auth and request fields)
	guardErr := guardEncryptCipherRequest(request, peerSPIFFEID, w, r)
	if guardErr != nil {
		return guardErr
	}

	// Get cipher only after auth passes
	c, cipherErr := getCipher()
	if cipherErr != nil {
		return cipherErr
	}

	nonce, ciphertext, encryptErr := encryptDataStreaming(plaintext, c, w)
	if encryptErr != nil {
		return encryptErr
	}

	return respondStreamingEncrypt(nonce, ciphertext, w)
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
	// Extract and validate SPIFFE ID before accessing cipher
	peerSPIFFEID, err := extractAndValidateSPIFFEID(w, r)
	if err != nil {
		return err
	}

	// Parse request (doesn't need cipher)
	request, jsonErr := readJSONEncryptRequestWithoutGuard(w, r)
	if jsonErr != nil {
		return jsonErr
	}

	// Full guard validation (auth and request fields)
	guardErr := guardEncryptCipherRequest(*request, peerSPIFFEID, w, r)
	if guardErr != nil {
		return guardErr
	}

	// Get cipher only after auth passes
	c, cipherErr := getCipher()
	if cipherErr != nil {
		return cipherErr
	}

	nonce, ciphertext, encryptErr := encryptDataJSON(
		request.Plaintext, c, w,
	)
	if encryptErr != nil {
		return encryptErr
	}

	return respondJSONEncrypt(nonce, ciphertext, w)
}
