//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package cipher

import (
	"net/http"

	"github.com/spiffe/spike-sdk-go/api/entity/data"
	"github.com/spiffe/spike-sdk-go/api/entity/v1/reqres"
	"github.com/spiffe/spike-sdk-go/log"
	"github.com/spiffe/spike/internal/net"
)

// respondStreamingDecrypt sends the decrypted plaintext as raw binary data
// for streaming mode requests.
//
// Parameters:
//   - plaintext: The decrypted data to send
//   - w: The HTTP response writer
//   - fName: The function name for logging
//
// Returns:
//   - error: An error if the response fails to send
func respondStreamingDecrypt(
	plaintext []byte, w http.ResponseWriter, fName string,
) error {
	w.Header().Set("Content-Type", "application/octet-stream")
	if _, err := w.Write(plaintext); err != nil {
		return err
	}
	log.Log().Info(fName, "message", "streaming decryption successful")
	return nil
}

// respondJSONDecrypt sends the decrypted plaintext as a structured JSON
// response for JSON mode requests.
//
// Parameters:
//   - plaintext: The decrypted data to send
//   - w: The HTTP response writer
//   - fName: The function name for logging
//
// Returns:
//   - error: An error if the response fails to send
func respondJSONDecrypt(
	plaintext []byte, w http.ResponseWriter, fName string,
) error {
	net.Success(
		reqres.CipherDecryptResponse{
			Plaintext: plaintext,
			Err:       data.ErrSuccess,
		}, w, fName,
	)
	return nil
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
//   - fName: The function name for logging
//
// Returns:
//   - error: An error if the response fails to send
func respondStreamingEncrypt(
	nonce, ciphertext []byte, w http.ResponseWriter, fName string,
) error {
	w.Header().Set("Content-Type", headerValueOctetStream)
	if _, err := w.Write([]byte{spikeCipherVersion}); err != nil {
		return err
	}
	if _, err := w.Write(nonce); err != nil {
		return err
	}
	if _, err := w.Write(ciphertext); err != nil {
		return err
	}
	log.Log().Info(fName, "message", "Streaming encryption successful")
	return nil
}

// respondJSONEncrypt sends the encrypted ciphertext as a structured JSON
// response for JSON mode requests.
//
// Parameters:
//   - nonce: The nonce bytes
//   - ciphertext: The encrypted data to send
//   - w: The HTTP response writer
//   - fName: The function name for logging
//
// Returns:
//   - error: An error if the response fails to send
func respondJSONEncrypt(
	nonce, ciphertext []byte, w http.ResponseWriter, fName string,
) error {
	net.Success(
		reqres.CipherEncryptResponse{
			Version:    spikeCipherVersion,
			Nonce:      nonce,
			Ciphertext: ciphertext,
			Err:        data.ErrSuccess,
		}, w, fName,
	)
	return nil
}
