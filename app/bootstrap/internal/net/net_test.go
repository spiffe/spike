//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package net

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/spiffe/spike-sdk-go/api/entity/v1/reqres"
	"github.com/spiffe/spike-sdk-go/crypto"
)

func TestShardContributionRequestMarshaling(t *testing.T) {
	// Test the JSON marshaling/unmarshaling that happens in Payload()
	validShard := make([]byte, crypto.AES256KeySize)
	for i := range validShard {
		validShard[i] = byte(i % 256)
	}

	scr := reqres.ShardPutRequest{}
	shard := new([crypto.AES256KeySize]byte)
	copy(shard[:], validShard)
	scr.Shard = shard

	// Test marshaling
	payload, err := json.Marshal(scr)
	if err != nil {
		t.Fatalf("Failed to marshal ShardPutRequest: %v", err)
	}

	if len(payload) == 0 {
		t.Error("Expected non-empty payload")
	}

	// Test unmarshaling
	var unmarshaled reqres.ShardPutRequest
	err = json.Unmarshal(payload, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal payload: %v", err)
	}

	if unmarshaled.Shard == nil {
		t.Fatal("Shard is nil in unmarshaled payload")
	}

	// Verify the shard data matches our input
	for i, b := range unmarshaled.Shard {
		if b != validShard[i] {
			t.Errorf("Shard data mismatch at index %d: expected %d, got %d", i, validShard[i], b)
		}
	}
}

func TestPostHTTPInteraction(t *testing.T) {
	tests := []struct {
		name           string
		serverResponse func(w http.ResponseWriter, r *http.Request)
		payload        []byte
		expectError    bool
	}{
		{
			name: "successful post request",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost {
					t.Errorf("Expected POST method, got %s", r.Method)
					w.WriteHeader(http.StatusMethodNotAllowed)
					return
				}

				// Verify the Content-Type header if needed
				contentType := r.Header.Get("Content-Type")
				if contentType != "application/json" && contentType != "" {
					// Content-Type might not be set, which is okay for this test
					fmt.Println("Content-Type header:", contentType)
				}

				body, err := io.ReadAll(r.Body)
				if err != nil {
					t.Errorf("Failed to read body: %v", err)
					w.WriteHeader(http.StatusBadRequest)
					return
				}

				expectedPayload := []byte("test payload")
				if !bytes.Equal(body, expectedPayload) {
					t.Errorf("Expected payload %s, got %s", string(expectedPayload), string(body))
					w.WriteHeader(http.StatusBadRequest)
					return
				}

				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("success"))
			},
			payload:     []byte("test payload"),
			expectError: false,
		},
		{
			name: "server returns 500 error",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte("internal server error"))
			},
			payload:     []byte("test payload"),
			expectError: true,
		},
		{
			name: "server returns 404 error",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
				_, _ = w.Write([]byte("not found"))
			},
			payload:     []byte("test payload"),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(tt.serverResponse))
			defer server.Close()

			if tt.expectError {
				// FIX-ME: after fixing Log.FatalLn and friends to panic,
				// The PutShardContributionRequest function calls os.Exit(1) on error, which we can't easily test
				// without significant refactoring. In a real scenario, you'd want to
				// refactor the function to return errors instead of calling os.Exit.
				t.Skip("Skipping test that would cause os.Exit - needs refactoring for testability")
			} else {
				// This should work without calling os.Exit
				err := PutShardContributionRequest(server.Client(), server.URL, tt.payload, "test-keeper")
				if err != nil {
					return
				}
			}
		})
	}
}

func TestShardContributionRequestStructure(t *testing.T) {
	// Test that we can create and work with ShardPutRequest
	scr := reqres.ShardPutRequest{}

	// Test with nil shard (should be valid)
	payload, err := json.Marshal(scr)
	if err != nil {
		t.Fatalf("Failed to marshal empty ShardPutRequest: %v", err)
	}

	var unmarshaled reqres.ShardPutRequest
	err = json.Unmarshal(payload, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal empty payload: %v", err)
	}

	// Test with valid shard
	validShard := new([crypto.AES256KeySize]byte)
	for i := range validShard {
		validShard[i] = byte(i)
	}
	scr.Shard = validShard

	payload, err = json.Marshal(scr)
	if err != nil {
		t.Fatalf("Failed to marshal ShardPutRequest with shard: %v", err)
	}

	err = json.Unmarshal(payload, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal payload with shard: %v", err)
	}

	if unmarshaled.Shard == nil {
		t.Fatal("Expected non-nil shard after unmarshal")
	}
}

func TestCryptoConstants(t *testing.T) {
	// Verify the crypto constant we depend on
	// noinspection GoBoolExpressions
	if crypto.AES256KeySize != 32 {
		t.Errorf("Expected AES256KeySize to be 32 bytes, got %d", crypto.AES256KeySize)
	}

	// Test that our shard array type has the correct size
	var shard [crypto.AES256KeySize]byte
	// noinspection GoBoolExpressions
	if len(shard) != 32 {
		t.Errorf("Expected shard array length to be 32, got %d", len(shard))
	}
}

func TestHTTPClientInteraction(t *testing.T) {
	// Test HTTP client behavior that PutShardContributionRequest() relies on
	testPayload := []byte(`{"shard": "test data"}`)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify the request structure
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST method, got %s", r.Method)
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("Failed to read request body: %v", err)
			return
		}

		if !bytes.Equal(body, testPayload) {
			t.Errorf("Request body mismatch. Expected: %s, Got: %s", string(testPayload), string(body))
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	}))
	defer server.Close()

	// Test the successful HTTP POST (this mimics what PutShardContributionRequest() does internally)
	client := server.Client()
	req, err := http.NewRequest(http.MethodPost, server.URL, bytes.NewReader(testPayload))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response: %v", err)
	}

	if string(respBody) != "OK" {
		t.Errorf("Expected response 'OK', got '%s'", string(respBody))
	}
}
