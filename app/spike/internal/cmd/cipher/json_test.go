//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package cipher

import (
	"encoding/base64"
	"strings"
	"testing"

	"github.com/spiffe/spike-sdk-go/api/entity/data"
)

// TestMarshalEncrypted_MatchesTheJSONDocument pins the exact bytes of the
// encrypt output: one JSON object naming the version, the nonce, and the
// ciphertext, byte fields base64-encoded by encoding/json, terminated by a
// single newline. Anything else breaks either operators reading the output
// or "decrypt --input" consuming it.
func TestMarshalEncrypted_MatchesTheJSONDocument(t *testing.T) {
	doc, err := marshalEncrypted(&data.EncryptedData{
		Version:    1,
		Nonce:      []byte("abc"),
		Ciphertext: []byte("xyz"),
	})
	if err != nil {
		t.Fatalf("marshalEncrypted() error = %v, want nil", err)
	}

	// base64("abc") and base64("xyz") have no padding; the expected
	// document is written out in full to pin the key order and shape.
	const want = "{\"version\":1,\"nonce\":\"YWJj\",\"ciphertext\":\"eHl6\"}\n"
	if string(doc) != want {
		t.Errorf("marshalEncrypted() = %q, want %q", string(doc), want)
	}
}

// TestParseEncryptedJSON checks the --input parsing paths: the happy path
// and every rejection. Every rejection document supplies all three keys
// unless the row's name says a key is missing or null, and each row
// asserts the distinctive text of the failure it names, so a row cannot
// pass on a different error than the one it tests.
func TestParseEncryptedJSON(t *testing.T) {
	gotVersion, gotNonce, gotCiphertext, err := parseEncryptedJSON(
		`{"version":7,"nonce":"YWJj","ciphertext":"eHl6"}`,
	)
	if err != nil {
		t.Fatalf("parseEncryptedJSON() error = %v, want nil", err)
	}
	if gotVersion != "7" {
		t.Errorf("parseEncryptedJSON() version = %q, want %q", gotVersion, "7")
	}
	if gotNonce != "YWJj" {
		t.Errorf("parseEncryptedJSON() nonce = %q, want %q", gotNonce, "YWJj")
	}
	if gotCiphertext != "eHl6" {
		t.Errorf("parseEncryptedJSON() ciphertext = %q, want %q",
			gotCiphertext, "eHl6")
	}

	// Version 0 is a real value: a present, non-null zero must parse.
	zeroVersion, _, _, zeroErr := parseEncryptedJSON(
		`{"version":0,"nonce":"YWJj","ciphertext":"eHl6"}`,
	)
	if zeroErr != nil {
		t.Fatalf("parseEncryptedJSON() version 0 error = %v, want nil", zeroErr)
	}
	if zeroVersion != "0" {
		t.Errorf("parseEncryptedJSON() version 0 = %q, want %q", zeroVersion, "0")
	}

	// Rows whose error comes from encoding/json share the "invalid --input
	// JSON" prefix with every other row, so detail also checks the error
	// that encoding/json wrapped under it.
	tests := []struct {
		name   string
		input  string
		want   string
		detail func(error) bool
	}{
		{
			"not JSON", "version=1",
			"invalid --input JSON",
			isSyntaxError,
		},
		{
			"empty object", `{}`,
			`"version" is missing or null`,
			nil,
		},
		{
			"null document", `null`,
			`"version" is missing or null`,
			nil,
		},
		{
			"missing nonce", `{"version":1}`,
			`"nonce" is missing or null`,
			nil,
		},
		{
			"missing ciphertext", `{"version":1,"nonce":"YWJj"}`,
			`"ciphertext" is missing or null`,
			nil,
		},
		{
			"null version",
			`{"version":null,"nonce":"YQ==","ciphertext":"YQ=="}`,
			`"version" is missing or null`,
			nil,
		},
		{
			"null nonce",
			`{"version":1,"nonce":null,"ciphertext":"YQ=="}`,
			`"nonce" is missing or null`,
			nil,
		},
		{
			"null ciphertext",
			`{"version":1,"nonce":"YQ==","ciphertext":null}`,
			`"ciphertext" is missing or null`,
			nil,
		},
		{
			"version out of range",
			`{"version":256,"nonce":"YQ==","ciphertext":"YQ=="}`,
			"invalid --input JSON",
			typeErrorOn("version"),
		},
		{
			"version wrong type",
			`{"version":"1","nonce":"YQ==","ciphertext":"YQ=="}`,
			"invalid --input JSON",
			typeErrorOn("version"),
		},
		{
			"nonce wrong type",
			`{"version":1,"nonce":7,"ciphertext":"YQ=="}`,
			"invalid --input JSON",
			typeErrorOn("nonce"),
		},
		// encoding/json does not name the field in a base64 error, so these
		// two rows can only check the kind of error.
		{
			"nonce not base64",
			`{"version":1,"nonce":"%%%","ciphertext":"YQ=="}`,
			"invalid --input JSON",
			isBase64Error,
		},
		{
			"ciphertext not base64",
			`{"version":1,"nonce":"YQ==","ciphertext":"%%%"}`,
			"invalid --input JSON",
			isBase64Error,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, _, gotErr := parseEncryptedJSON(tt.input)
			if gotErr == nil {
				t.Fatal("parseEncryptedJSON() error = nil, want an error")
			}
			if !strings.Contains(gotErr.Error(), tt.want) {
				t.Errorf("parseEncryptedJSON() error = %q, want it to contain %q",
					gotErr.Error(), tt.want)
			}
			if tt.detail != nil && !tt.detail(gotErr) {
				t.Errorf("parseEncryptedJSON() error = %q, want the "+
					"encoding/json error wrapped under it", gotErr.Error())
			}
		})
	}
}

// TestEncryptedJSON_RoundTrips feeds the marshalEncrypted output to
// parseEncryptedJSON and asserts that the components survive: the format
// encrypt prints is the document decrypt --input accepts. The nonce and
// ciphertext are chosen to need base64 padding, so the canonical padded
// form is exercised in both directions.
func TestEncryptedJSON_RoundTrips(t *testing.T) {
	encrypted := &data.EncryptedData{
		Version:    1,
		Nonce:      []byte{0x01, 0x02},
		Ciphertext: []byte("hello"),
	}

	doc, err := marshalEncrypted(encrypted)
	if err != nil {
		t.Fatalf("marshalEncrypted() error = %v, want nil", err)
	}

	gotVersion, gotNonce, gotCiphertext, err := parseEncryptedJSON(string(doc))
	if err != nil {
		t.Fatalf("parseEncryptedJSON() error = %v, want nil", err)
	}

	if want := "1"; gotVersion != want {
		t.Errorf("round-trip version = %q, want %q", gotVersion, want)
	}
	if want := base64.StdEncoding.EncodeToString(encrypted.Nonce); gotNonce != want {
		t.Errorf("round-trip nonce = %q, want %q", gotNonce, want)
	}
	if want := base64.StdEncoding.EncodeToString(encrypted.Ciphertext); gotCiphertext != want {
		t.Errorf("round-trip ciphertext = %q, want %q", gotCiphertext, want)
	}
}

// TestDecrypt_InputIsExclusiveWithOtherSources drives the decrypt command
// through Cobra the way the root command does and asserts the conflict
// errors, so mixing --input with a component flag or --file fails loudly
// instead of resolving per-field in ways the user cannot see.
func TestDecrypt_InputIsExclusiveWithOtherSources(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{
			"input with version",
			[]string{"--input", `{"version":1}`, "--version", "1"},
			"--input and --version cannot be combined",
		},
		{
			"input with nonce",
			[]string{"--input", `{"version":1}`, "--nonce", "YWJj"},
			"--input and --nonce cannot be combined",
		},
		{
			"input with ciphertext",
			[]string{"--input", `{"version":1}`, "--ciphertext", "eHl6"},
			"--input and --ciphertext cannot be combined",
		},
		{
			"input with file",
			[]string{"--input", `{"version":1}`, "--file", "some-file"},
			"--input and --file cannot be combined",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := newSilencedRoot(newDecryptCommand(nil, pilotSPIFFEID()))
			root.SetOut(&strings.Builder{})
			root.SetErr(&strings.Builder{})
			root.SetArgs(append([]string{"decrypt"}, tt.args...))

			execErr := root.Execute()
			if execErr == nil {
				t.Fatal("Execute() error = nil, want a conflict error")
			}
			if execErr.Error() != tt.want {
				t.Errorf("Execute() error = %q, want %q",
					execErr.Error(), tt.want)
			}
		})
	}
}

// TestDecrypt_InputPairsWithAlgorithm locks the one flag --input does
// combine with: the document does not carry the algorithm, so --algorithm
// must keep working next to --input. Driven with a nil source, the
// expected error is the unavailable source, proving the combination got
// past every conflict check.
func TestDecrypt_InputPairsWithAlgorithm(t *testing.T) {
	root := newSilencedRoot(newDecryptCommand(nil, pilotSPIFFEID()))
	root.SetOut(&strings.Builder{})
	root.SetErr(&strings.Builder{})
	root.SetArgs([]string{
		"decrypt",
		"--input", `{"version":1,"nonce":"YWJj","ciphertext":"eHl6"}`,
		"--algorithm", "AES-GCM",
	})

	execErr := root.Execute()
	if execErr == nil {
		t.Fatal("Execute() error = nil, want the source error")
	}
	const want = "SPIFFE X509 source is unavailable"
	if execErr.Error() != want {
		t.Errorf("Execute() error = %q, want %q", execErr.Error(), want)
	}
}

// TestDecrypt_EmptyInputIsRejected covers an --input that is present but
// empty, as a failed command substitution produces. It must fail as an
// invalid document rather than count as absent and fall through to stream
// mode, which would wait on stdin. Driven with a nil source, the fallthrough
// would surface as the unavailable source instead.
func TestDecrypt_EmptyInputIsRejected(t *testing.T) {
	root := newSilencedRoot(newDecryptCommand(nil, pilotSPIFFEID()))
	root.SetOut(&strings.Builder{})
	root.SetErr(&strings.Builder{})
	root.SetArgs([]string{"decrypt", "--input", ""})

	execErr := root.Execute()
	if execErr == nil {
		t.Fatal("Execute() error = nil, want an invalid --input error")
	}
	if !strings.HasPrefix(execErr.Error(), "invalid --input JSON") ||
		!isSyntaxError(execErr) {
		t.Errorf("Execute() error = %q, want an invalid --input JSON "+
			"syntax error", execErr.Error())
	}
}
