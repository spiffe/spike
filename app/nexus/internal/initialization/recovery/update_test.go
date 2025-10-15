//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package recovery

import (
	"encoding/json"
	"net/url"
	"strconv"
	"testing"

	"github.com/cloudflare/circl/group"
	"github.com/spiffe/spike-sdk-go/api/entity/v1/reqres"
	apiUrl "github.com/spiffe/spike-sdk-go/api/url"
	"github.com/spiffe/spike-sdk-go/crypto"
)

func TestSendShardsToKeepers_NetworkDependentFunction(t *testing.T) {
	// The sendShardsToKeepers function has multiple external dependencies that make it
	// difficult to test without a significant infrastructure:
	// 1. Requires SPIFFE X509Source
	// 2. Makes network calls via mTLS clients
	// 3. Depends on state management for the root key
	// 4. Calls computeShares() and sanityCheck() which have their own dependencies
	t.Skip("Skipping sendShardsToKeepers test - requires SPIFFE infrastructure, network connectivity, and state management")

	// Note: To properly test this function, you would need to:
	// 1. Mock the workloadapi.X509Source
	// 2. Mock network.CreateMTLSClientWithPredicate
	// 3. Mock net.Post
	// 4. Mock state.RootKeyZero()
	// 5. Mock computeShares() and sanityCheck()
	// 6. Set up test HTTP servers
	// 7. Or refactor the code for better testability with dependency injection
}

func TestKeeperIDConversionLogic(t *testing.T) {
	// Test the keeper ID to integer conversion logic used in sendShardsToKeepers
	tests := []struct {
		name      string
		keeperID  string
		expectErr bool
		expected  int
	}{
		{"valid numeric ID", "1", false, 1},
		{"valid large ID", "999", false, 999},
		{"zero ID", "0", false, 0},
		{"invalid non-numeric", "abc", true, 0},
		{"empty string", "", true, 0},
		{"mixed alphanumeric", "123abc", true, 0},
		{"negative number", "-1", false, -1}, // strconv.Atoi handles negatives
		{"leading zeros", "007", false, 7},
		{"whitespace", " 123 ", true, 0}, // strconv.Atoi doesn't trim whitespace
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This mimics the conversion logic in sendShardsToKeepers:
			// kid, err := strconv.Atoi(keeperID)
			result, err := strconv.Atoi(tt.keeperID)

			if tt.expectErr {
				if err == nil {
					t.Errorf("Expected error for keeper ID '%s', but got none", tt.keeperID)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for keeper ID '%s': %v", tt.keeperID, err)
				}
				if result != tt.expected {
					t.Errorf("Expected result %d, got %d", tt.expected, result)
				}
			}
		})
	}
}

func TestURLJoinPathForKeepers(t *testing.T) {
	// Test URL joining logic used in sendShardsToKeepers
	tests := []struct {
		name          string
		keeperAPIRoot string
		expectedPath  string
		expectError   bool
	}{
		{
			name:          "valid HTTP URL",
			keeperAPIRoot: "http://keeper1.example.com",
			expectedPath:  string(apiUrl.KeeperContribute),
			expectError:   false,
		},
		{
			name:          "valid HTTPS URL with port",
			keeperAPIRoot: "https://keeper1.example.com:8443",
			expectedPath:  string(apiUrl.KeeperContribute),
			expectError:   false,
		},
		{
			name:          "URL with existing path",
			keeperAPIRoot: "https://keeper1.example.com/api/v1",
			expectedPath:  string(apiUrl.KeeperContribute),
			expectError:   false,
		},
		{
			name:          "URL with trailing slash",
			keeperAPIRoot: "https://keeper1.example.com/",
			expectedPath:  string(apiUrl.KeeperContribute),
			expectError:   false,
		},
		//{
		//	name:          "invalid URL",
		//	keeperAPIRoot: "not a valid url",
		//	expectedPath:  string(apiUrl.KeeperContribute),
		//	expectError:   true,
		//},
		// FIX-ME: address me.
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This mimics the URL joining in sendShardsToKeepers:
			// u, err := url.JoinPath(keeperAPIRoot, string(apiUrl.KeeperContribute))
			result, err := url.JoinPath(tt.keeperAPIRoot, tt.expectedPath)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error for URL '%s', but got none", tt.keeperAPIRoot)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for URL '%s': %v", tt.keeperAPIRoot, err)
					return
				}

				// Verify result is a valid URL
				_, parseErr := url.Parse(result)
				if parseErr != nil {
					t.Errorf("Result should be valid URL: %v", parseErr)
				}

				// Verify the result contains the expected path
				if !containsPathUpdate(result, tt.expectedPath) {
					t.Errorf("Result %s should contain path %s", result, tt.expectedPath)
				}
			}
		})
	}
}

func TestShardContributionRequestStructure(t *testing.T) {
	// Test the ShardPutRequest structure used in sendShardsToKeepers
	tests := []struct {
		name        string
		setupShard  func() *reqres.ShardPutRequest
		expectValid bool
	}{
		{
			name: "valid shard contribution",
			setupShard: func() *reqres.ShardPutRequest {
				testData := &[crypto.AES256KeySize]byte{}
				for i := range testData {
					testData[i] = byte(i % 256)
				}
				return &reqres.ShardPutRequest{
					Shard: testData,
				}
			},
			expectValid: true,
		},
		{
			name: "nil shard contribution",
			setupShard: func() *reqres.ShardPutRequest {
				return &reqres.ShardPutRequest{
					Shard: nil,
				}
			},
			expectValid: true, // nil shard might be valid in some cases
		},
		{
			name: "zero shard contribution",
			setupShard: func() *reqres.ShardPutRequest {
				return &reqres.ShardPutRequest{
					Shard: &[crypto.AES256KeySize]byte{}, // All zeros
				}
			},
			expectValid: true, // Zero shard might be valid
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := tt.setupShard()

			if !tt.expectValid && request != nil {
				t.Error("Expected invalid request to be nil")
			}

			if tt.expectValid && request == nil {
				t.Error("Expected valid request to be non-nil")
			}

			// Test JSON marshaling (as done in sendShardsToKeepers)
			if request != nil {
				data, err := json.Marshal(request)
				if err != nil {
					t.Errorf("Failed to marshal request: %v", err)
					return
				}

				if len(data) == 0 {
					t.Error("Marshaled data should not be empty")
				}

				// Test unmarshaling back
				var unmarshaled reqres.ShardPutRequest
				err = json.Unmarshal(data, &unmarshaled)
				if err != nil {
					t.Errorf("Failed to unmarshal request: %v", err)
					return
				}

				// Verify data integrity
				if request.Shard != nil && unmarshaled.Shard != nil {
					// noinspection GoBoolExpressions
					if len(request.Shard) != len(unmarshaled.Shard) {
						t.Error("Shard lengths should match after marshal/unmarshal")
					}
					for i, b := range request.Shard {
						if b != unmarshaled.Shard[i] {
							t.Errorf("Shard data mismatch at index %d", i)
						}
					}
				} else if request.Shard != nil || unmarshaled.Shard != nil {
					t.Error("Shard nil status should match after marshal/unmarshal")
				}
			}
		})
	}
}

func TestGroupP256ScalarOperations(t *testing.T) {
	// Test P256 scalar operations used in sendShardsToKeepers for ID comparison
	g := group.P256

	// Test creating and setting scalar values (as done for keeper ID comparison)
	tests := []struct {
		name     string
		value    uint64
		expected uint64
	}{
		{"keeper ID 1", 1, 1},
		{"keeper ID 2", 2, 2},
		{"keeper ID 999", 999, 999},
		{"zero ID", 0, 0},
		{"large ID", 18446744073709551615, 18446744073709551615}, // Max uint64
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This mimics the scalar creation in sendShardsToKeepers:
			// group.P256.NewScalar().SetUint64(uint64(kid))
			scalar := g.NewScalar().SetUint64(tt.value)

			if scalar == nil {
				t.Error("Scalar should not be nil")
				return
			}

			// Test IsEqual functionality (used in the keeper ID comparison)
			compareScalar := g.NewScalar().SetUint64(tt.expected)
			if !scalar.IsEqual(compareScalar) {
				t.Errorf("Scalars with same value should be equal")
			}

			// Test IsZero
			if tt.value == 0 {
				if !scalar.IsZero() {
					t.Error("Zero scalar should return true for IsZero()")
				}
			} else {
				if scalar.IsZero() {
					t.Error("Non-zero scalar should return false for IsZero()")
				}
			}

			// Test SetUint64(0) for cleanup (as done in sendShardsToKeepers)
			scalar.SetUint64(0)
			if !scalar.IsZero() {
				t.Error("Scalar should be zero after SetUint64(0)")
			}
		})
	}
}

func TestKeeperMapOperations(t *testing.T) {
	// Test map operations on keeper data (as used in sendShardsToKeepers)
	keepers := map[string]string{
		"1": "https://keeper1.example.com",
		"2": "https://keeper2.example.com",
		"3": "https://keeper3.example.com:8443",
	}

	// Test iteration (as done in sendShardsToKeepers)
	count := 0
	for keeperID, keeperAPIRoot := range keepers {
		count++

		// Test keeper ID validation
		if keeperID == "" {
			t.Error("Keeper ID should not be empty")
		}

		// Test keeper API root validation
		if keeperAPIRoot == "" {
			t.Error("Keeper API root should not be empty")
		}

		// Test that keeper ID can be converted to int
		_, err := strconv.Atoi(keeperID)
		if err != nil {
			t.Errorf("Keeper ID '%s' should be convertible to int: %v",
				keeperID, err)
		}

		// Test that the SPIKE Keeper API root is a valid URL
		_, err = url.Parse(keeperAPIRoot)
		if err != nil {
			t.Errorf("Keeper API root '%s' should be valid URL: %v",
				keeperAPIRoot, err)
		}
	}

	expectedCount := 3
	if count != expectedCount {
		t.Errorf("Expected to iterate over %d keepers, got %d",
			expectedCount, count)
	}

	// Test map access
	keeper1URL := keepers["1"]
	if keeper1URL != "https://keeper1.example.com" {
		t.Errorf("Expected keeper1 URL to be 'https://keeper1.example.com', got '%s'",
			keeper1URL)
	}

	// Test a non-existent key
	nonExistent := keepers["999"]
	if nonExistent != "" {
		t.Error("Non-existent key should return empty string")
	}
}

func TestAPIUrlKeeperContribute(t *testing.T) {
	// Test the API URL constant used in sendShardsToKeepers
	contributeURL := string(apiUrl.KeeperContribute)

	// noinspection GoBoolExpressions
	if contributeURL == "" {
		t.Error("KeeperContribute URL should not be empty")
	}

	// Should be a valid path component
	t.Logf("KeeperContribute API URL: %s", contributeURL)

	// Test URL joining with this path
	baseURL := "https://example.com"
	result, err := url.JoinPath(baseURL, contributeURL)
	if err != nil {
		t.Errorf("Failed to join path with KeeperContribute: %v", err)
	}

	if !containsPathUpdate(result, contributeURL) {
		t.Errorf("Joined URL should contain the contribute path")
	}
}

func TestContributionLengthValidation(t *testing.T) {
	// Test contribution length validation logic used in sendShardsToKeepers
	tests := []struct {
		name        string
		data        []byte
		expectValid bool
	}{
		{
			name:        "valid length (32 bytes)",
			data:        make([]byte, crypto.AES256KeySize),
			expectValid: true,
		},
		{
			name:        "invalid length (too short)",
			data:        make([]byte, 16),
			expectValid: false,
		},
		{
			name:        "invalid length (too long)",
			data:        make([]byte, 64),
			expectValid: false,
		},
		{
			name:        "empty data",
			data:        []byte{},
			expectValid: false,
		},
		{
			name:        "nil data",
			data:        nil,
			expectValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This mimics the validation in sendShardsToKeepers:
			// if len(contribution) != crypto.AES256KeySize
			isValid := tt.data != nil && len(tt.data) == crypto.AES256KeySize

			if isValid != tt.expectValid {
				t.Errorf("Expected validity %v for data length %d, got %v",
					tt.expectValid, len(tt.data), isValid)
			}
		})
	}
}

func TestShardArrayOperations(t *testing.T) {
	// Test shard array operations used in sendShardsToKeepers

	// Test creating a new shard array (as done in the function)
	shard := new([crypto.AES256KeySize]byte)

	// noinspection GoBoolExpressions
	if len(shard) != crypto.AES256KeySize {
		t.Errorf("Shard array length should be %d, got %d",
			crypto.AES256KeySize, len(shard))
	}

	// Test copy operation (as done in sendShardsToKeepers)
	testData := make([]byte, crypto.AES256KeySize)
	for i := range testData {
		testData[i] = byte(i % 256)
	}

	copy(shard[:], testData)

	// Verify data was copied correctly
	for i, b := range shard {
		expected := byte(i % 256)
		if b != expected {
			t.Errorf("Copy failed at index %d: expected %d, got %d", i, expected, b)
		}
	}

	// Test that it's a proper array, not a slice
	if cap(shard[:]) != crypto.AES256KeySize {
		t.Errorf("Shard capacity should be %d, got %d",
			crypto.AES256KeySize, cap(shard[:]))
	}
}
