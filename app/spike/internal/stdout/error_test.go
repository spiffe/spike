//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package stdout

import (
	"bytes"
	"testing"

	"github.com/spf13/cobra"
	sdkErrors "github.com/spiffe/spike-sdk-go/errors"
)

// createTestCommand creates a Cobra command with a captured stderr buffer.
func createTestCommand(commandPath string) (*cobra.Command, *bytes.Buffer) {
	buf := &bytes.Buffer{}
	cmd := &cobra.Command{
		Use: commandPath,
	}
	cmd.SetErr(buf)
	return cmd, buf
}

// createTestCommandWithParent creates a Cobra command hierarchy for testing
// command group detection (e.g., "spike cipher encrypt").
func createTestCommandWithParent(group, subcommand string) (*cobra.Command, *bytes.Buffer) {
	buf := &bytes.Buffer{}

	root := &cobra.Command{Use: "spike"}
	groupCmd := &cobra.Command{Use: group}
	subCmd := &cobra.Command{Use: subcommand}

	root.AddCommand(groupCmd)
	groupCmd.AddCommand(subCmd)
	subCmd.SetErr(buf)

	return subCmd, buf
}

func TestHandleAPIError_NilError(t *testing.T) {
	cmd, _ := createTestCommand("test")

	result := HandleAPIError(cmd, nil)

	if result {
		t.Error("HandleAPIError(nil) = true, want false")
	}
}

func TestHandleAPIError_CommonErrors(t *testing.T) {
	tests := []struct {
		name        string
		err         *sdkErrors.SDKError
		wantMessage string
	}{
		{
			name:        "ErrDataMarshalFailure",
			err:         sdkErrors.ErrDataMarshalFailure,
			wantMessage: "Error: Malformed request.",
		},
		{
			name:        "ErrDataUnmarshalFailure",
			err:         sdkErrors.ErrDataUnmarshalFailure,
			wantMessage: "Error: Failed to parse API response.",
		},
		{
			name:        "ErrAPINotFound",
			err:         sdkErrors.ErrAPINotFound,
			wantMessage: "Error: Resource not found.",
		},
		{
			name:        "ErrAPIBadRequest",
			err:         sdkErrors.ErrAPIBadRequest,
			wantMessage: "Error: Invalid request.",
		},
		{
			name:        "ErrDataInvalidInput",
			err:         sdkErrors.ErrDataInvalidInput,
			wantMessage: "Error: Invalid input provided.",
		},
		{
			name:        "ErrNetPeerConnection",
			err:         sdkErrors.ErrNetPeerConnection,
			wantMessage: "Error: Failed to connect to SPIKE Nexus.",
		},
		{
			name:        "ErrAccessUnauthorized",
			err:         sdkErrors.ErrAccessUnauthorized,
			wantMessage: "Error: Unauthorized access.",
		},
		{
			name:        "ErrNetReadingResponseBody",
			err:         sdkErrors.ErrNetReadingResponseBody,
			wantMessage: "Error: Failed to read response body.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, buf := createTestCommand("test")

			result := HandleAPIError(cmd, tt.err)

			if !result {
				t.Error("HandleAPIError() = false, want true")
			}

			if !bytes.Contains(buf.Bytes(), []byte(tt.wantMessage)) {
				t.Errorf("HandleAPIError() output = %q, want to contain %q",
					buf.String(), tt.wantMessage)
			}
		})
	}
}

func TestHandleAPIError_PolicyErrors(t *testing.T) {
	tests := []struct {
		name        string
		err         *sdkErrors.SDKError
		wantMessage string
	}{
		{
			name:        "ErrEntityNotFound",
			err:         sdkErrors.ErrEntityNotFound,
			wantMessage: "Error: Entity not found.",
		},
		{
			name:        "ErrEntityInvalid",
			err:         sdkErrors.ErrEntityInvalid,
			wantMessage: "Error: Invalid entity.",
		},
		{
			name:        "ErrAPIPostFailed",
			err:         sdkErrors.ErrAPIPostFailed,
			wantMessage: "Error: Operation failed.",
		},
		{
			name:        "ErrEntityCreationFailed",
			err:         sdkErrors.ErrEntityCreationFailed,
			wantMessage: "Error: Failed to create resource.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, buf := createTestCommandWithParent("policy", "create")

			result := HandleAPIError(cmd, tt.err)

			if !result {
				t.Error("HandleAPIError() = false, want true")
			}

			if !bytes.Contains(buf.Bytes(), []byte(tt.wantMessage)) {
				t.Errorf("HandleAPIError() output = %q, want to contain %q",
					buf.String(), tt.wantMessage)
			}
		})
	}
}

func TestHandleAPIError_CipherErrors(t *testing.T) {
	tests := []struct {
		name        string
		err         *sdkErrors.SDKError
		wantMessage string
	}{
		{
			name:        "ErrCryptoEncryptionFailed",
			err:         sdkErrors.ErrCryptoEncryptionFailed,
			wantMessage: "Error: Encryption operation failed.",
		},
		{
			name:        "ErrCryptoDecryptionFailed",
			err:         sdkErrors.ErrCryptoDecryptionFailed,
			wantMessage: "Error: Decryption operation failed.",
		},
		{
			name:        "ErrCryptoCipherNotAvailable",
			err:         sdkErrors.ErrCryptoCipherNotAvailable,
			wantMessage: "Error: Cipher not available.",
		},
		{
			name:        "ErrCryptoInvalidEncryptionKeyLength",
			err:         sdkErrors.ErrCryptoInvalidEncryptionKeyLength,
			wantMessage: "Error: Invalid encryption key length.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, buf := createTestCommandWithParent("cipher", "encrypt")

			result := HandleAPIError(cmd, tt.err)

			if !result {
				t.Error("HandleAPIError() = false, want true")
			}

			if !bytes.Contains(buf.Bytes(), []byte(tt.wantMessage)) {
				t.Errorf("HandleAPIError() output = %q, want to contain %q",
					buf.String(), tt.wantMessage)
			}
		})
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
			cmd, _ := createTestCommandWithParent(tt.group, tt.subCmd)

			result := getCommandGroup(cmd)

			if result != tt.expected {
				t.Errorf("getCommandGroup() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestGetCommandGroup_ShortPath(t *testing.T) {
	cmd := &cobra.Command{Use: "spike"}

	result := getCommandGroup(cmd)

	if result != "" {
		t.Errorf("getCommandGroup() = %q, want empty string", result)
	}
}

func TestHandlePolicyError_NonPolicyError(t *testing.T) {
	cmd, _ := createTestCommand("test")
	// Use an error that isn't a policy-specific error
	err := sdkErrors.ErrCryptoEncryptionFailed

	result := handlePolicyError(cmd, err)

	if result {
		t.Error("handlePolicyError() = true for non-policy error, want false")
	}
}

func TestHandleCipherError_NonCipherError(t *testing.T) {
	cmd, _ := createTestCommand("test")
	// Use an error that isn't a cipher-specific error
	err := sdkErrors.ErrEntityNotFound

	result := handleCipherError(cmd, err)

	if result {
		t.Error("handleCipherError() = true for non-cipher error, want false")
	}
}
