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
	"github.com/spiffe/spike-sdk-go/net"
)

// readJSONDecryptRequestWithoutGuard reads and parses a JSON mode decryption
// request without performing guard validation.
//
// Parameters:
//   - w: The HTTP response writer for error responses
//   - r: The HTTP request containing the JSON data
//
// Returns:
//   - *reqres.CipherDecryptRequest: The parsed request
//   - *sdkErrors.SDKError: An error if reading or parsing fails
func readJSONDecryptRequestWithoutGuard(
	w http.ResponseWriter, r *http.Request,
) (*reqres.CipherDecryptRequest, *sdkErrors.SDKError) {
	requestBody, err := net.ReadRequestBodyAndRespondOnFail(w, r)
	if err != nil {
		return nil, err
	}

	request, unmarshalErr := net.UnmarshalAndRespondOnFail[
		reqres.CipherDecryptRequest, reqres.CipherDecryptResponse](
		requestBody, w,
		reqres.CipherDecryptResponse{}.BadRequest(),
	)
	if unmarshalErr != nil {
		return nil, unmarshalErr
	}

	return request, nil
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
//   - *sdkErrors.SDKError: An error if reading fails
func readStreamingDecryptRequestData(
	w http.ResponseWriter, r *http.Request, c cipher.AEAD,
) (byte, []byte, []byte, *sdkErrors.SDKError) {
	const fName = "readStreamingDecryptRequestData"

	// Read the version byte
	ver := make([]byte, 1)
	n, err := io.ReadFull(r.Body, ver)
	if err != nil || n != 1 {
		failErr := sdkErrors.ErrCryptoFailedToReadVersion.Clone()
		log.WarnErr(fName, *failErr)
		http.Error(
			w, string(failErr.Code), http.StatusBadRequest,
		)
		return 0, nil, nil, failErr
	}

	version := ver[0]

	// Validate version matches the expected value
	if version != spikeCipherVersion {
		failErr := sdkErrors.ErrCryptoUnsupportedCipherVersion.Clone()
		log.WarnErr(fName, *failErr)
		http.Error(
			w, string(failErr.Code), http.StatusBadRequest,
		)
		return 0, nil, nil, failErr
	}

	// Read the nonce
	bytesToRead := c.NonceSize()
	nonce := make([]byte, bytesToRead)
	n, err = io.ReadFull(r.Body, nonce)
	if err != nil || n != bytesToRead {
		failErr := sdkErrors.ErrCryptoFailedToReadNonce.Clone()
		log.WarnErr(fName, *failErr)
		http.Error(
			w, string(failErr.Code), http.StatusBadRequest,
		)
		return 0, nil, nil, failErr
	}

	// Read the remaining body as ciphertext
	ciphertext, readErr := io.ReadAll(r.Body)
	if readErr != nil {
		failErr := sdkErrors.ErrDataReadFailure.Wrap(readErr)
		failErr.Msg = "failed to read ciphertext"
		log.WarnErr(fName, *failErr)
		http.Error(
			w, string(failErr.Code), http.StatusBadRequest,
		)
		return 0, nil, nil, failErr
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
//   - *sdkErrors.SDKError: An error if reading fails
func readStreamingEncryptRequestWithoutGuard(
	w http.ResponseWriter, r *http.Request,
) ([]byte, *sdkErrors.SDKError) {
	plaintext, err := net.ReadRequestBodyAndRespondOnFail(w, r)
	if err != nil {
		return nil, err
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
//   - *reqres.CipherEncryptRequest: The parsed request
//   - *sdkErrors.SDKError: An error if reading or parsing fails
func readJSONEncryptRequestWithoutGuard(
	w http.ResponseWriter, r *http.Request,
) (*reqres.CipherEncryptRequest, *sdkErrors.SDKError) {
	requestBody, err := net.ReadRequestBodyAndRespondOnFail(w, r)
	if err != nil {
		return nil, err
	}

	request, unmarshalErr := net.UnmarshalAndRespondOnFail[
		reqres.CipherEncryptRequest, reqres.CipherEncryptResponse](
		requestBody, w,
		reqres.CipherEncryptResponse{}.BadRequest(),
	)
	if unmarshalErr != nil {
		return nil, unmarshalErr
	}

	return request, nil
}
