//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package cipher

import (
	"crypto/cipher"
	"net/http"

	"github.com/spiffe/spike-sdk-go/api/entity/v1/reqres"
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
//   - fName: The function name for logging
//
// Returns:
//   - error: An error if any step fails
func handleStreamingDecrypt(
	w http.ResponseWriter, r *http.Request,
	getCipher func() (cipher.AEAD, error), fName string,
) error {
	// NOTE: since we are dealing with streaming data, we cannot directly use
	// the request parameter validation patterns that we employ in the JSON/REST
	// payloads. We need to read the entire stream and generate a request
	// entity accordingly.

	// Extract and validate SPIFFE ID before accessing cipher
	peerSPIFFEID, err := extractAndValidateSPIFFEID(w, r)
	if err != nil {
		return err
	}

	// Get cipher only after SPIFFE ID validation passes
	c, err := getCipher()
	if err != nil {
		return err
	}

	// Read request data (now that we have cipher for nonce size)
	version, nonce, ciphertext, err := readStreamingDecryptRequestData(
		w, r, c,
	)
	if err != nil {
		return err
	}

	// Construct request object for guard validation
	request := reqres.CipherDecryptRequest{
		Version:    version,
		Nonce:      nonce,
		Ciphertext: ciphertext,
	}

	// Full guard validation (auth + request fields)
	err = guardDecryptCipherRequest(request, peerSPIFFEID, w, r)
	if err != nil {
		return err
	}

	plaintext, err := decryptData(nonce, ciphertext, c, w, true, fName)
	if err != nil {
		return err
	}

	return respondStreamingDecrypt(plaintext, w, fName)
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
//   - fName: The function name for logging
//
// Returns:
//   - error: An error if any step fails
func handleJSONDecrypt(
	w http.ResponseWriter, r *http.Request,
	getCipher func() (cipher.AEAD, error), fName string,
) error {
	// Extract and validate SPIFFE ID before accessing cipher
	peerSPIFFEID, err := extractAndValidateSPIFFEID(w, r)
	if err != nil {
		return err
	}

	// Parse request (doesn't need cipher)
	request, err := readJSONDecryptRequestWithoutGuard(w, r)
	if err != nil {
		return err
	}

	// Full guard validation (auth + request fields)
	err = guardDecryptCipherRequest(request, peerSPIFFEID, w, r)
	if err != nil {
		return err
	}

	// Get cipher only after auth passes
	c, err := getCipher()
	if err != nil {
		return err
	}

	err = validateJSONDecryptData(request.Version, request.Nonce, c, w)
	if err != nil {
		return err
	}

	plaintext, err := decryptData(
		request.Nonce, request.Ciphertext, c, w, false, fName,
	)
	if err != nil {
		return err
	}

	return respondJSONDecrypt(plaintext, w, fName)
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
//   - fName: The function name for logging
//
// Returns:
//   - error: An error if any step fails
func handleStreamingEncrypt(
	w http.ResponseWriter, r *http.Request,
	getCipher func() (cipher.AEAD, error), fName string,
) error {
	// Extract and validate SPIFFE ID before accessing cipher
	peerSPIFFEID, err := extractAndValidateSPIFFEID(w, r)
	if err != nil {
		return err
	}

	// Read plaintext (doesn't need cipher)
	plaintext, err := readStreamingEncryptRequestWithoutGuard(w, r)
	if err != nil {
		return err
	}

	// Construct request object for guard validation
	request := reqres.CipherEncryptRequest{
		Plaintext: plaintext,
	}

	// Full guard validation (auth + request fields)
	err = guardEncryptCipherRequest(request, peerSPIFFEID, w, r)
	if err != nil {
		return err
	}

	// Get cipher only after auth passes
	c, err := getCipher()
	if err != nil {
		return err
	}

	nonce, ciphertext, err := encryptData(plaintext, c, w, true, fName)
	if err != nil {
		return err
	}

	return respondStreamingEncrypt(nonce, ciphertext, w, fName)
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
//   - fName: The function name for logging
//
// Returns:
//   - error: An error if any step fails
func handleJSONEncrypt(
	w http.ResponseWriter, r *http.Request,
	getCipher func() (cipher.AEAD, error), fName string,
) error {
	// Extract and validate SPIFFE ID before accessing cipher
	peerSPIFFEID, err := extractAndValidateSPIFFEID(w, r)
	if err != nil {
		return err
	}

	// Parse request (doesn't need cipher)
	request, err := readJSONEncryptRequestWithoutGuard(w, r)
	if err != nil {
		return err
	}

	// Full guard validation (auth + request fields)
	err = guardEncryptCipherRequest(request, peerSPIFFEID, w, r)
	if err != nil {
		return err
	}

	// Get cipher only after auth passes
	c, err := getCipher()
	if err != nil {
		return err
	}

	nonce, ciphertext, err := encryptData(
		request.Plaintext, c, w, false, fName,
	)
	if err != nil {
		return err
	}

	return respondJSONEncrypt(nonce, ciphertext, w, fName)
}
