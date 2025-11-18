//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package cipher

import (
	"crypto/cipher"
	"io"
	"net/http"

	"github.com/spiffe/spike-sdk-go/api/entity/v1/reqres"
	sdkErrors "github.com/spiffe/spike-sdk-go/errors"
	"github.com/spiffe/spike-sdk-go/log"

	"github.com/spiffe/spike/internal/net"
)

// readJSONDecryptRequestWithoutGuard reads and parses a JSON mode decryption
// request without performing guard validation.
//
// Parameters:
//   - w: The HTTP response writer for error responses
//   - r: The HTTP request containing the JSON data
//
// Returns:
//   - reqres.CipherDecryptRequest: The parsed request
//   - error: An error if reading or parsing fails
func readJSONDecryptRequestWithoutGuard(
	w http.ResponseWriter, r *http.Request,
) (reqres.CipherDecryptRequest, error) {
	requestBody := net.ReadRequestBody(w, r)
	if requestBody == nil {
		return reqres.CipherDecryptRequest{}, sdkErrors.ErrReadFailure
	}

	request := net.HandleRequest[
		reqres.CipherDecryptRequest, reqres.CipherDecryptResponse](
		requestBody, w,
		reqres.CipherDecryptResponse{Err: sdkErrors.ErrCodeBadInput},
	)
	if request == nil {
		return reqres.CipherDecryptRequest{}, sdkErrors.ErrParseFailure
	}

	return *request, nil
}

// readStreamingDecryptRequestData reads the binary data from a streaming mode
// decryption request (version, nonce, ciphertext).
//
// This function does NOT perform authentication - the caller must have already
// called the guard function.
//
// The streaming format is: version byte + nonce and ciphertext
//
// Parameters:
//   - w: The HTTP response writer for error responses
//   - r: The HTTP request containing the binary data
//   - c: The cipher to determine nonce size
//
// Returns:
//   - version: The protocol version byte
//   - nonce: The nonce bytes
//   - ciphertext: The encrypted data
//   - error: An error if reading fails
func readStreamingDecryptRequestData(
	w http.ResponseWriter, r *http.Request, c cipher.AEAD,
) (byte, []byte, []byte, error) {
	const fName = "readStreamingDecryptRequestData"

	// Read version byte
	ver := make([]byte, 1)
	n, err := io.ReadFull(r.Body, ver)
	if err != nil || n != 1 {
		log.Log().Debug(fName, "message", sdkErrors.ErrCodeCryptoFailedToReadVersion)
		http.Error(
			w, string(sdkErrors.ErrCodeCryptoFailedToReadVersion), http.StatusBadRequest,
		)
		return 0, nil, nil, sdkErrors.ErrCryptoFailedToReadVersion
	}
	version := ver[0]

	// Read nonce
	bytesToRead := c.NonceSize()
	nonce := make([]byte, bytesToRead)
	n, err = io.ReadFull(r.Body, nonce)
	if err != nil || n != bytesToRead {
		log.Log().Debug(fName, "message", sdkErrors.ErrCodeCryptoFailedToReadNonce)
		http.Error(
			w, string(sdkErrors.ErrCodeCryptoFailedToReadNonce), http.StatusBadRequest,
		)
		return 0, nil, nil, sdkErrors.ErrCryptoFailedToReadNonce
	}

	// Read the remaining body as ciphertext
	ciphertext := net.ReadRequestBody(w, r)
	if ciphertext == nil {
		return 0, nil, nil, sdkErrors.ErrReadFailure
	}

	return version, nonce, ciphertext, nil
}

// readStreamingEncryptRequestWithoutGuard reads a streaming mode encryption
// request without performing guard validation.
//
// Parameters:
//   - w: The HTTP response writer for error responses
//   - r: The HTTP request containing the binary data
//
// Returns:
//   - plaintext: The plaintext data to encrypt
//   - error: An error if reading fails
func readStreamingEncryptRequestWithoutGuard(
	w http.ResponseWriter, r *http.Request,
) ([]byte, error) {
	plaintext := net.ReadRequestBody(w, r)
	if plaintext == nil {
		return nil, sdkErrors.ErrReadFailure
	}

	return plaintext, nil
}

// readJSONEncryptRequestWithoutGuard reads and parses a JSON mode encryption
// request without performing guard validation.
//
// Parameters:
//   - w: The HTTP response writer for error responses
//   - r: The HTTP request containing the JSON data
//
// Returns:
//   - reqres.CipherEncryptRequest: The parsed request
//   - error: An error if reading or parsing fails
func readJSONEncryptRequestWithoutGuard(
	w http.ResponseWriter, r *http.Request,
) (reqres.CipherEncryptRequest, error) {
	requestBody := net.ReadRequestBody(w, r)
	if requestBody == nil {
		return reqres.CipherEncryptRequest{}, sdkErrors.ErrReadFailure
	}

	request := net.HandleRequest[
		reqres.CipherEncryptRequest, reqres.CipherEncryptResponse](
		requestBody, w, reqres.CipherEncryptBadInput,
	)
	if request == nil {
		return reqres.CipherEncryptRequest{}, sdkErrors.ErrParseFailure
	}

	return *request, nil
}
