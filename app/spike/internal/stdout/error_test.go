//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package stdout

import (
	"testing"

	"github.com/spf13/cobra"
	sdkErrors "github.com/spiffe/spike-sdk-go/errors"
)

func TestAPIError_NilError(t *testing.T) {
	if got := APIError(createTestCommand("test"), nil); got != nil {
		t.Errorf("APIError(nil) = %v, want nil", got)
	}
}

func TestAPIError_CommonErrors(t *testing.T) {
	tests := []struct {
		name        string
		err         *sdkErrors.SDKError
		wantMessage string
	}{
		{
			name:        "ErrDataMarshalFailure",
			err:         sdkErrors.ErrDataMarshalFailure,
			wantMessage: "malformed request",
		},
		{
			name:        "ErrDataUnmarshalFailure",
			err:         sdkErrors.ErrDataUnmarshalFailure,
			wantMessage: "failed to parse API response",
		},
		{
			name:        "ErrAPINotFound",
			err:         sdkErrors.ErrAPINotFound,
			wantMessage: "resource not found",
		},
		{
			name:        "ErrAPIBadRequest",
			err:         sdkErrors.ErrAPIBadRequest,
			wantMessage: "invalid request",
		},
		{
			name:        "ErrDataInvalidInput",
			err:         sdkErrors.ErrDataInvalidInput,
			wantMessage: "invalid input provided",
		},
		{
			name:        "ErrNetPeerConnection",
			err:         sdkErrors.ErrNetPeerConnection,
			wantMessage: "failed to connect to SPIKE Nexus",
		},
		{
			name:        "ErrAccessUnauthorized",
			err:         sdkErrors.ErrAccessUnauthorized,
			wantMessage: "unauthorized access",
		},
		{
			name:        "ErrNetReadingResponseBody",
			err:         sdkErrors.ErrNetReadingResponseBody,
			wantMessage: "failed to read response body",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := APIError(createTestCommand("test"), tt.err)
			if got == nil {
				t.Fatal("APIError() = nil, want an error")
			}
			if got.Error() != tt.wantMessage {
				t.Errorf("APIError() = %q, want %q",
					got.Error(), tt.wantMessage)
			}
		})
	}
}

func TestAPIError_PolicyErrors(t *testing.T) {
	tests := []struct {
		name        string
		err         *sdkErrors.SDKError
		wantMessage string
	}{
		{
			name:        "ErrEntityNotFound",
			err:         sdkErrors.ErrEntityNotFound,
			wantMessage: "entity not found",
		},
		{
			name:        "ErrEntityInvalid",
			err:         sdkErrors.ErrEntityInvalid,
			wantMessage: "invalid entity",
		},
		{
			name:        "ErrAPIPostFailed",
			err:         sdkErrors.ErrAPIPostFailed,
			wantMessage: "operation failed",
		},
		{
			name:        "ErrEntityCreationFailed",
			err:         sdkErrors.ErrEntityCreationFailed,
			wantMessage: "failed to create resource",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := createTestCommandWithParent("policy", "create")
			got := APIError(cmd, tt.err)
			if got == nil {
				t.Fatal("APIError() = nil, want an error")
			}
			if got.Error() != tt.wantMessage {
				t.Errorf("APIError() = %q, want %q",
					got.Error(), tt.wantMessage)
			}
		})
	}
}

func TestAPIError_CipherErrors(t *testing.T) {
	tests := []struct {
		name        string
		err         *sdkErrors.SDKError
		wantMessage string
	}{
		{
			name:        "ErrCryptoEncryptionFailed",
			err:         sdkErrors.ErrCryptoEncryptionFailed,
			wantMessage: "encryption operation failed",
		},
		{
			name:        "ErrCryptoDecryptionFailed",
			err:         sdkErrors.ErrCryptoDecryptionFailed,
			wantMessage: "decryption operation failed",
		},
		{
			name:        "ErrCryptoCipherNotAvailable",
			err:         sdkErrors.ErrCryptoCipherNotAvailable,
			wantMessage: "cipher not available",
		},
		{
			name:        "ErrCryptoInvalidEncryptionKeyLength",
			err:         sdkErrors.ErrCryptoInvalidEncryptionKeyLength,
			wantMessage: "invalid encryption key length",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := createTestCommandWithParent("cipher", "encrypt")
			got := APIError(cmd, tt.err)
			if got == nil {
				t.Fatal("APIError() = nil, want an error")
			}
			if got.Error() != tt.wantMessage {
				t.Errorf("APIError() = %q, want %q",
					got.Error(), tt.wantMessage)
			}
		})
	}
}

func TestAPIError_FallbackReturnsTheSDKError(t *testing.T) {
	// A cipher error outside the cipher command group is not mapped, so the
	// SDK error itself must reach the caller.
	cmd := createTestCommandWithParent("secret", "get")
	got := APIError(cmd, sdkErrors.ErrCryptoEncryptionFailed)
	if got == nil {
		t.Fatal("APIError() = nil, want an error")
	}
	if got.Error() != sdkErrors.ErrCryptoEncryptionFailed.Error() {
		t.Errorf("APIError() = %q, want the SDK error text %q",
			got.Error(), sdkErrors.ErrCryptoEncryptionFailed.Error())
	}
}

func TestGetCommandGroup(t *testing.T) {
	tests := []struct {
		name     string
		group    string
		subCmd   string
		expected string
	}{
		{"cipher group", "cipher", "encrypt", "cipher"},
		{"policy group", "policy", "create", "policy"},
		{"secret group", "secret", "get", "secret"},
		{"operator group", "operator", "status", "operator"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := createTestCommandWithParent(tt.group, tt.subCmd)
			if got := getCommandGroup(cmd); got != tt.expected {
				t.Errorf("getCommandGroup() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestGetCommandGroup_ShortPath(t *testing.T) {
	if got := getCommandGroup(&cobra.Command{Use: "spike"}); got != "" {
		t.Errorf("getCommandGroup() = %q, want an empty string", got)
	}
}

func TestPolicyError_NonPolicyError(t *testing.T) {
	if got := policyError(sdkErrors.ErrCryptoEncryptionFailed); got != nil {
		t.Errorf("policyError() = %v for a non-policy error, want nil", got)
	}
}

func TestCipherError_NonCipherError(t *testing.T) {
	if got := cipherError(sdkErrors.ErrEntityNotFound); got != nil {
		t.Errorf("cipherError() = %v for a non-cipher error, want nil", got)
	}
}
