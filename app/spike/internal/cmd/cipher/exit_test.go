//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package cipher

import (
	"bytes"
	"strings"
	"testing"
)

// TestDecrypt_FailsThroughCobraWithoutSource drives the decrypt command the
// way the root command does and asserts that a failing precondition surfaces
// as a returned error, which is what makes the CLI exit non-zero.
func TestDecrypt_FailsThroughCobraWithoutSource(t *testing.T) {
	root := newSilencedRoot(newDecryptCommand(nil, pilotSPIFFEID()))
	stdoutBuf := &bytes.Buffer{}
	stderrBuf := &bytes.Buffer{}
	root.SetOut(stdoutBuf)
	root.SetErr(stderrBuf)
	root.SetArgs([]string{"decrypt", "--version", "999"})

	execErr := root.Execute()
	if execErr == nil {
		t.Fatal("Execute() error = nil, want an error without an X509 source")
	}
	const want = "SPIFFE X509 source is unavailable"
	if execErr.Error() != want {
		t.Errorf("Execute() error = %q, want %q", execErr.Error(), want)
	}
	if stdoutBuf.Len() != 0 {
		t.Errorf("stdout = %q, want nothing on failure", stdoutBuf.String())
	}
	if stderrBuf.Len() != 0 {
		t.Errorf("stderr = %q, want nothing: the root command prints once",
			stderrBuf.String())
	}
}

// TestEncrypt_FailsThroughCobraWithoutSource is the encrypt counterpart.
func TestEncrypt_FailsThroughCobraWithoutSource(t *testing.T) {
	root := newSilencedRoot(newEncryptCommand(nil, pilotSPIFFEID()))
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"encrypt", "--plaintext", "aGVsbG8="})

	if execErr := root.Execute(); execErr == nil {
		t.Fatal("Execute() error = nil, want an error without an X509 source")
	}
}

// TestDecryptJSON_RejectsBadInputsBeforeCallingTheAPI checks the JSON-mode
// validation paths. The API client is nil on purpose: every case must fail
// before the client is touched.
func TestDecryptJSON_RejectsBadInputsBeforeCallingTheAPI(t *testing.T) {
	tests := []struct {
		name       string
		version    string
		nonce      string
		ciphertext string
		want       string
	}{
		{"version not a number", "x", "", "", "invalid --version, must be 0-255"},
		{"version too large", "256", "", "", "invalid --version, must be 0-255"},
		{"version negative", "-1", "", "", "invalid --version, must be 0-255"},
		{"bad nonce", "1", "%%%", "", "invalid --nonce base64"},
		{"bad ciphertext", "1", "YQ==", "%%%", "invalid --ciphertext base64"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotErr := decryptJSON(
				nil, tt.version, tt.nonce, tt.ciphertext, "", "",
			)
			if gotErr == nil {
				t.Fatal("decryptJSON() error = nil, want an error")
			}
			if gotErr.Error() != tt.want {
				t.Errorf("decryptJSON() error = %q, want %q",
					gotErr.Error(), tt.want)
			}
		})
	}
}

// TestEncryptJSON_RejectsBadPlaintext checks the base64 validation path.
func TestEncryptJSON_RejectsBadPlaintext(t *testing.T) {
	gotErr := encryptJSON(nil, "%%%", "", "")
	if gotErr == nil {
		t.Fatal("encryptJSON() error = nil, want an error")
	}
	const want = "invalid --plaintext base64"
	if gotErr.Error() != want {
		t.Errorf("encryptJSON() error = %q, want %q", gotErr.Error(), want)
	}
}

// TestStream_MissingInputFileIsAnError checks that a missing --file fails
// before any API call.
func TestStream_MissingInputFileIsAnError(t *testing.T) {
	missing := t.TempDir() + "/does-not-exist"

	decErr := decryptStream(nil, missing, "")
	if decErr == nil ||
		!strings.Contains(decErr.Error(), "input file does not exist") {
		t.Errorf("decryptStream() error = %v, want input file does not exist",
			decErr)
	}

	encErr := encryptStream(nil, missing, "")
	if encErr == nil ||
		!strings.Contains(encErr.Error(), "input file does not exist") {
		t.Errorf("encryptStream() error = %v, want input file does not exist",
			encErr)
	}
}
