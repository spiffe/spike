//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package cipher

import (
	"net/http"

	"github.com/spiffe/spike-sdk-go/api/entity/v1/reqres"
	sdkErrors "github.com/spiffe/spike-sdk-go/errors"
	"github.com/spiffe/spike-sdk-go/net"
)

// respondStreamingDecrypt sends the decrypted plaintext as raw binary data
// for streaming mode requests.
//
// Parameters:
//   - plaintext: The decrypted data to send
//   - w: The HTTP response writer
//
// Returns:
//   - *sdkErrors.SDKError: An error if the response fails to send, nil on
//     success
func respondStreamingDecrypt(
	plaintext []byte, w http.ResponseWriter,
) *sdkErrors.SDKError {
	w.Header().Set(headerKeyContentType, headerValueOctetStream)
	if _, err := w.Write(plaintext); err != nil {
		return sdkErrors.ErrFSStreamWriteFailed.Wrap(err)
	}
	return nil
}

// respondJSONDecrypt sends the decrypted plaintext as a structured JSON
// response for JSON mode requests.
//
// Parameters:
//   - plaintext: The decrypted data to send
//   - w: The HTTP response writer
//
// Returns:
//   - *sdkErrors.SDKError: nil on success, or an error if the response fails
//     to send
func respondJSONDecrypt(
	plaintext []byte, w http.ResponseWriter,
) *sdkErrors.SDKError {
	return net.Success(
		reqres.CipherDecryptResponse{
			Plaintext: plaintext,
		}.Success(), w,
	)
}

// respondStreamingEncrypt sends the encrypted ciphertext as raw binary data
// for streaming mode requests.
//
// The streaming format is: version byte + nonce + ciphertext
//
// Parameters:
//   - nonce: The nonce bytes
//   - ciphertext: The encrypted data to send
//   - w: The HTTP response writer
//
// Returns:
//   - *sdkErrors.SDKError: An error if the response fails to send, nil on
//     success
func respondStreamingEncrypt(
	nonce, ciphertext []byte, w http.ResponseWriter,
) *sdkErrors.SDKError {
	w.Header().Set(headerKeyContentType, headerValueOctetStream)
	if _, err := w.Write([]byte{spikeCipherVersion}); err != nil {
		return sdkErrors.ErrFSStreamWriteFailed.Wrap(err)
	}
	if _, err := w.Write(nonce); err != nil {
		return sdkErrors.ErrFSStreamWriteFailed.Wrap(err)
	}
	if _, err := w.Write(ciphertext); err != nil {
		return sdkErrors.ErrFSStreamWriteFailed.Wrap(err)
	}
	return nil
}

// respondJSONEncrypt sends the encrypted ciphertext as a structured JSON
// response for JSON mode requests.
//
// Parameters:
//   - nonce: The nonce bytes
//   - ciphertext: The encrypted data to send
//   - w: The HTTP response writer
//
// Returns:
//   - *sdkErrors.SDKError: nil on success, or an error if the response fails
//     to send
func respondJSONEncrypt(
	nonce, ciphertext []byte, w http.ResponseWriter,
) *sdkErrors.SDKError {
	return net.Success(
		reqres.CipherEncryptResponse{
			Version:    spikeCipherVersion,
			Nonce:      nonce,
			Ciphertext: ciphertext,
		}.Success(), w,
	)
}
