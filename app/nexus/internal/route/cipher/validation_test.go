//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package cipher

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

type testErrorResponse struct {
	Err string `json:"err"`
}

func TestValidateVersion_ValidVersion(t *testing.T) {
	w := httptest.NewRecorder()
	errResp := testErrorResponse{Err: "invalid version"}

	err := validateVersion(spikeCipherVersion, w, errResp, "test")

	if err != nil {
		t.Errorf("validateVersion() error = %v, want nil", err)
	}

	if w.Code != http.StatusOK {
		t.Errorf("validateVersion() status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestValidateVersion_InvalidVersion(t *testing.T) {
	w := httptest.NewRecorder()
	errResp := testErrorResponse{Err: "invalid version"}

	err := validateVersion(byte('9'), w, errResp, "test")

	if err == nil {
		t.Error("validateVersion() expected error for invalid version")
	}

	if w.Code != http.StatusBadRequest {
		t.Errorf("validateVersion() status = %d, want %d",
			w.Code, http.StatusBadRequest)
	}
}

func TestValidateNonceSize_ValidNonce(t *testing.T) {
	w := httptest.NewRecorder()
	errResp := testErrorResponse{Err: "invalid nonce"}
	// expectedNonceSize is 12 bytes for AES-GCM
	nonce := make([]byte, expectedNonceSize)

	err := validateNonceSize(nonce, w, errResp, "test")

	if err != nil {
		t.Errorf("validateNonceSize() error = %v, want nil", err)
	}

	if w.Code != http.StatusOK {
		t.Errorf("validateNonceSize() status = %d, want %d",
			w.Code, http.StatusOK)
	}
}

func TestValidateNonceSize_InvalidNonce(t *testing.T) {
	tests := []struct {
		name      string
		nonceSize int
	}{
		{"too short", 8},
		{"too long", 16},
		{"empty", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			errResp := testErrorResponse{Err: "invalid nonce"}
			nonce := make([]byte, tt.nonceSize)

			err := validateNonceSize(nonce, w, errResp, "test")

			if err == nil {
				t.Error("validateNonceSize() expected error for invalid nonce size")
			}

			if w.Code != http.StatusBadRequest {
				t.Errorf("validateNonceSize() status = %d, want %d",
					w.Code, http.StatusBadRequest)
			}
		})
	}
}

func TestValidateCiphertextSize_ValidSize(t *testing.T) {
	tests := []struct {
		name string
		size int
	}{
		{"small", 100},
		{"medium", 1000},
		{"at max", maxCiphertextSize},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			errResp := testErrorResponse{Err: "ciphertext too large"}
			ciphertext := make([]byte, tt.size)

			err := validateCiphertextSize(ciphertext, w, errResp, "test")

			if err != nil {
				t.Errorf("validateCiphertextSize() error = %v, want nil", err)
			}

			if w.Code != http.StatusOK {
				t.Errorf("validateCiphertextSize() status = %d, want %d",
					w.Code, http.StatusOK)
			}
		})
	}
}

func TestValidateCiphertextSize_TooLarge(t *testing.T) {
	w := httptest.NewRecorder()
	errResp := testErrorResponse{Err: "ciphertext too large"}
	ciphertext := make([]byte, maxCiphertextSize+1)

	err := validateCiphertextSize(ciphertext, w, errResp, "test")

	if err == nil {
		t.Error("validateCiphertextSize() expected error for oversized ciphertext")
	}

	if w.Code != http.StatusBadRequest {
		t.Errorf("validateCiphertextSize() status = %d, want %d",
			w.Code, http.StatusBadRequest)
	}
}

func TestValidatePlaintextSize_ValidSize(t *testing.T) {
	tests := []struct {
		name string
		size int
	}{
		{"small", 100},
		{"medium", 1000},
		{"at max", maxPlaintextSize},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			errResp := testErrorResponse{Err: "plaintext too large"}
			plaintext := make([]byte, tt.size)

			err := validatePlaintextSize(plaintext, w, errResp, "test")

			if err != nil {
				t.Errorf("validatePlaintextSize() error = %v, want nil", err)
			}

			if w.Code != http.StatusOK {
				t.Errorf("validatePlaintextSize() status = %d, want %d",
					w.Code, http.StatusOK)
			}
		})
	}
}

func TestValidatePlaintextSize_TooLarge(t *testing.T) {
	w := httptest.NewRecorder()
	errResp := testErrorResponse{Err: "plaintext too large"}
	plaintext := make([]byte, maxPlaintextSize+1)

	err := validatePlaintextSize(plaintext, w, errResp, "test")

	if err == nil {
		t.Error("validatePlaintextSize() expected error for oversized plaintext")
	}

	if w.Code != http.StatusBadRequest {
		t.Errorf("validatePlaintextSize() status = %d, want %d",
			w.Code, http.StatusBadRequest)
	}
}

func TestCipherConstants(t *testing.T) {
	// Verify AES-GCM standard values
	if expectedNonceSize != 12 {
		t.Errorf("expectedNonceSize = %d, want 12 (AES-GCM standard)",
			expectedNonceSize)
	}

	// maxPlaintextSize should be 16 bytes less than maxCiphertextSize
	// to account for the AES-GCM authentication tag
	if maxPlaintextSize != maxCiphertextSize-16 {
		t.Errorf("maxPlaintextSize = %d, want %d (maxCiphertextSize - 16)",
			maxPlaintextSize, maxCiphertextSize-16)
	}

	// Verify version byte
	if spikeCipherVersion != byte('1') {
		t.Errorf("spikeCipherVersion = %v, want '1'", spikeCipherVersion)
	}
}
