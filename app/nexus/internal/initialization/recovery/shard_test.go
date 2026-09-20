//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package recovery

import (
	"encoding/json"
	"net/url"
	"testing"

	"github.com/spiffe/spike-sdk-go/api/entity/v1/reqres"
	apiUrl "github.com/spiffe/spike-sdk-go/api/url"
	"github.com/spiffe/spike-sdk-go/crypto"
	sdkErrors "github.com/spiffe/spike-sdk-go/errors"
	"github.com/spiffe/spike-sdk-go/security/mem"
)

func TestShardURL_ValidInput(t *testing.T) {
	tests := []struct {
		name           string
		keeperAPIRoot  string
		expectedSuffix string
	}{
		{
			name:           "valid HTTPS URL",
			keeperAPIRoot:  "https://example.com",
			expectedSuffix: string(apiUrl.KeeperShard),
		},
		{
			name:           "URL with port",
			keeperAPIRoot:  "https://example.com:8443",
			expectedSuffix: string(apiUrl.KeeperShard),
		},
		{
			name:           "URL with path",
			keeperAPIRoot:  "https://example.com/api/v1",
			expectedSuffix: string(apiUrl.KeeperShard),
		},
		{
			name:           "URL ending with slash",
			keeperAPIRoot:  "https://example.com/",
			expectedSuffix: string(apiUrl.KeeperShard),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, urlErr := keeperURL(tt.keeperAPIRoot, tt.expectedSuffix)
			if urlErr != nil {
				t.Fatalf("Unexpected error for %q: %v", tt.keeperAPIRoot, urlErr)
				return
			}
			if result == "" {
				t.Error("Expected non-empty result")
				return
			}

			// Verify the result contains the keeper API root
			if !containsBase(result, tt.keeperAPIRoot) {
				t.Errorf("Result %s should contain base URL %s",
					result, tt.keeperAPIRoot)
			}

			// Verify the result contains the keeper shard path
			if !containsPath(result, tt.expectedSuffix) {
				t.Errorf("Result %s should contain path %s",
					result, tt.expectedSuffix)
			}

			// Verify it's a valid URL
			if _, parseErr := url.Parse(result); parseErr != nil {
				t.Errorf("Result should be valid URL: %v", parseErr)
			}
		})
	}
}

func TestShardURL_InvalidInput(t *testing.T) {
	tests := []struct {
		name          string
		keeperAPIRoot string
	}{
		{
			name:          "invalid URL with spaces",
			keeperAPIRoot: "https://example .com",
		},
		{
			name:          "invalid URL with newline",
			keeperAPIRoot: "https://example.com\n",
		},
		{
			name:          "empty string",
			keeperAPIRoot: "",
		},
		{
			name:          "whitespace only",
			keeperAPIRoot: "   ",
		},
		{
			name:          "plain http is not allowed for mTLS",
			keeperAPIRoot: "http://example.com",
		},
		{
			name:          "relative path without a scheme",
			keeperAPIRoot: "not a url",
		},
		{
			name:          "scheme without a host",
			keeperAPIRoot: "https://",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, urlErr := keeperURL(
				tt.keeperAPIRoot, string(apiUrl.KeeperShard),
			)
			if urlErr == nil {
				t.Fatalf("Expected an error for %q, got URL %q",
					tt.keeperAPIRoot, result)
				return
			}
			if urlErr.Code != sdkErrors.ErrDataInvalidInput.Code {
				t.Errorf("Expected error code %q, got %q",
					sdkErrors.ErrDataInvalidInput.Code, urlErr.Code)
			}
			if result != "" {
				t.Errorf("Expected an empty URL on error, got %q", result)
			}
		})
	}
}

func TestUnmarshalShardResponse_ValidInput(t *testing.T) {
	// Create a valid ShardGetResponse for testing
	testShard := &[crypto.AES256KeySize]byte{}
	for i := range testShard {
		testShard[i] = byte(i % 256)
	}

	validResponse := reqres.ShardGetResponse{
		Shard: testShard,
	}

	// Marshal it to JSON
	data, err := json.Marshal(validResponse)
	if err != nil {
		t.Fatalf("Failed to marshal test data: %v", err)
	}

	// Test unmarshaling
	result, uerr := unmarshalShardResponse(data)

	if uerr != nil {
		t.Fatal("Expected non-nil result")
	}

	if result.Shard == nil {
		t.Fatal("Expected non-nil shard in result")
	}

	// Verify the data matches
	for i, b := range result.Shard {
		expected := byte(i % 256)
		if b != expected {
			t.Errorf("Data mismatch at index %d: expected %d, got %d",
				i, expected, b)
		}
	}
}

func TestUnmarshalShardResponse_InvalidInput(t *testing.T) {
	tests := []struct {
		name string
		data []byte
	}{
		{
			name: "empty data",
			data: []byte{},
		},
		{
			name: "invalid JSON",
			data: []byte("not json"),
		},
		{
			name: "malformed JSON",
			data: []byte("{invalid json}"),
		},
		{
			name: "null JSON",
			data: []byte("null"),
		},
		{
			name: "wrong structure",
			data: []byte(`{"wrong": "structure"}`),
		},
		{
			name: "keeper error code",
			data: []byte(`{"Err": "not_found"}`),
		},
		{
			name: "trailing data",
			data: []byte(`{"shard": null} {}`),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, unmarshalErr := unmarshalShardResponse(tt.data)

			// Invalid input should return an error and a nil result
			if unmarshalErr == nil {
				t.Error("Expected an error for invalid input")
			}
			if result != nil {
				t.Errorf("Expected nil result for invalid input, got %+v",
					result)
			}
		})
	}
}

func TestShardResponse_NetworkDependentFunction(t *testing.T) {
	// The ShardGetResponse function makes network calls and requires SPIFFE
	// infrastructure
	// We skip this test since it would hang or fail without proper setup
	t.Skip("Skipping ShardGetResponse test - requires SPIFFE source and " +
		"network connectivity")

	// Note: To properly test this function, you would need to:
	// 1. Mock the network.CreateMTLSClientWithPredicate function
	// 2. Mock the net.Post function
	// 3. Set up test SPIFFE infrastructure
	// 4. Create test HTTP servers
	// 5. Or refactor the code for better testability with dependency injection
}

func TestShardRequestMarshaling(t *testing.T) {
	// Test the ShardRequest marshaling used in ShardGetResponse
	shardRequest := reqres.ShardGetRequest{}

	data, err := json.Marshal(shardRequest)
	if err != nil {
		t.Errorf("Failed to marshal ShardRequest: %v", err)
	}

	if len(data) == 0 {
		t.Error("Marshaled data should not be empty")
	}

	// Test that it can be unmarshaled back
	var unmarshaled reqres.ShardGetRequest
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Errorf("Failed to unmarshal ShardRequest: %v", err)
	}

	// ShardRequest might be an empty struct, so just verify the process works
	t.Logf("Successfully marshaled and unmarshaled ShardRequest: %s",
		string(data))
}

func TestShardResponseStructure(t *testing.T) {
	// Test ShardGetResponse structure operations
	tests := []struct {
		name        string
		setupShard  func() *reqres.ShardGetResponse
		expectValid bool
	}{
		{
			name: "valid shard response",
			setupShard: func() *reqres.ShardGetResponse {
				testData := &[crypto.AES256KeySize]byte{}
				testData[0] = 42
				return &reqres.ShardGetResponse{
					Shard: testData,
				}
			},
			expectValid: true,
		},
		{
			name: "nil shard response",
			setupShard: func() *reqres.ShardGetResponse {
				return &reqres.ShardGetResponse{
					Shard: nil,
				}
			},
			expectValid: true, // nil shard might be valid in some cases
		},
		{
			name: "zero shard response",
			setupShard: func() *reqres.ShardGetResponse {
				return &reqres.ShardGetResponse{
					Shard: &[crypto.AES256KeySize]byte{}, // All zeros
				}
			},
			expectValid: true, // Zero shard might be valid
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := tt.setupShard()

			if !tt.expectValid && response != nil {
				t.Error("Expected invalid response to be nil")
			}

			if tt.expectValid && response == nil {
				t.Error("Expected valid response to be non-nil")
			}

			// Test JSON marshaling/unmarshaling
			if response != nil {
				data, err := json.Marshal(response)
				if err != nil {
					t.Errorf("Failed to marshal response: %v", err)
					return
				}

				var unmarshaled reqres.ShardGetResponse
				err = json.Unmarshal(data, &unmarshaled)
				if err != nil {
					t.Errorf("Failed to unmarshal response: %v", err)
					return
				}

				// Compare shard pointers (they won't be the same after
				// marshal/unmarshal)
				if response.Shard != nil && unmarshaled.Shard != nil {
					// noinspection GoBoolExpressions
					if len(response.Shard) != len(unmarshaled.Shard) {
						t.Error("Shard lengths should match after marshal/unmarshal")
					}
					for i, b := range response.Shard {
						if b != unmarshaled.Shard[i] {
							t.Errorf("Shard data mismatch at index %d", i)
						}
					}
				} else if response.Shard != nil || unmarshaled.Shard != nil {
					// One is nil, the other is not
					t.Error("Shard nil status should match after marshal/unmarshal")
				}
			}
		})
	}
}

func TestURLJoinPath(t *testing.T) {
	// Test the url.JoinPath functionality used in shardURL
	tests := []struct {
		name        string
		base        string
		path        string
		expectError bool
	}{
		{
			name:        "valid HTTP URL",
			base:        "http://example.com",
			path:        "api/shard",
			expectError: false,
		},
		{
			name:        "valid HTTPS URL",
			base:        "https://example.com",
			path:        "api/shard",
			expectError: false,
		},
		{
			name:        "URL with existing path",
			base:        "https://example.com/v1",
			path:        "api/shard",
			expectError: false,
		},
		{
			name:        "base with trailing slash",
			base:        "https://example.com/",
			path:        "api/shard",
			expectError: false,
		},
		{
			name:        "path with leading slash",
			base:        "https://example.com",
			path:        "api/shard",
			expectError: false,
		},
		{
			// An unterminated IPv6 literal cannot be parsed, so the join
			// fails. A bare string such as "not a url" is a valid relative
			// path for url.JoinPath and is rejected by keeperURL instead.
			name:        "invalid base URL",
			base:        "http://[::1",
			path:        "api/shard",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := url.JoinPath(tt.base, tt.path)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error for invalid input")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
					return
				}

				if result == "" {
					t.Error("Expected non-empty result")
					return
				}

				// Verify result is a valid URL
				_, parseErr := url.Parse(result)
				if parseErr != nil {
					t.Errorf("Result should be valid URL: %v", parseErr)
				}

				// Verify the result contains the path
				if !containsPath(result, tt.path) {
					t.Errorf("Result %s should contain path %s", result, tt.path)
				}
			}
		})
	}
}

func TestAPIUrlConstant(t *testing.T) {
	// Test that the API URL constant is properly defined
	keeperShard := string(apiUrl.KeeperShard)

	// noinspection GoBoolExpressions
	if keeperShard == "" {
		t.Error("KeeperShard API URL should not be empty")
	}

	// Should be a valid path component
	if keeperShard[0] == '/' {
		t.Log("KeeperShard starts with slash (absolute path)")
	} else {
		t.Log("KeeperShard is relative path")
	}

	t.Logf("KeeperShard API URL: %s", keeperShard)
}

func TestCryptoConstants(t *testing.T) {
	// Test that crypto constants are as expected
	// noinspection GoBoolExpressions
	if crypto.AES256KeySize != 32 {
		t.Errorf("Expected AES256KeySize to be 32, got %d", crypto.AES256KeySize)
	}

	// Test creating shard arrays
	var shard [crypto.AES256KeySize]byte
	// noinspection GoBoolExpressions
	if len(shard) != 32 {
		t.Errorf("Expected shard array length 32, got %d", len(shard))
	}

	// Test pointer to the shard array
	shardPtr := &shard
	// noinspection GoBoolExpressions
	if len(shardPtr) != 32 {
		t.Errorf("Expected shard pointer length 32, got %d", len(shardPtr))
	}
}

func TestResetShards_ZeroesAndEmpties(t *testing.T) {
	first := &[crypto.AES256KeySize]byte{1, 2, 3}
	second := &[crypto.AES256KeySize]byte{4, 5, 6}
	shards := map[string]*[crypto.AES256KeySize]byte{"1": first, "2": second}

	resetShards(shards)

	if len(shards) != 0 {
		t.Errorf("expected an empty map, got %d entries", len(shards))
	}
	if !mem.Zeroed32(first) || !mem.Zeroed32(second) {
		t.Error("expected the previous shard buffers to be zeroed")
	}
}
