//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package net

import (
	"bytes"
	"encoding/json"
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
	payload, marshalErr := json.Marshal(scr)
	if marshalErr != nil {
		t.Fatalf("Failed to marshal ShardPutRequest: %v", marshalErr)
	}

	if len(payload) == 0 {
		t.Error("Expected non-empty payload")
	}

	// Test unmarshaling
	var unmarshaled reqres.ShardPutRequest
	unmarshalErr := json.Unmarshal(payload, &unmarshaled)
	if unmarshalErr != nil {
		t.Fatalf("Failed to unmarshal payload: %v", unmarshalErr)
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

func TestShardContributionRequestStructure(t *testing.T) {
	// Test that we can create and work with ShardPutRequest
	scr := reqres.ShardPutRequest{}

	// Test with nil shard (should be valid)
	payload, marshalErr := json.Marshal(scr)
	if marshalErr != nil {
		t.Fatalf("Failed to marshal empty ShardPutRequest: %v", marshalErr)
	}

	var unmarshaled reqres.ShardPutRequest
	unmarshalErr := json.Unmarshal(payload, &unmarshaled)
	if unmarshalErr != nil {
		t.Fatalf("Failed to unmarshal empty payload: %v", unmarshalErr)
	}

	// Test with valid shard
	validShard := new([crypto.AES256KeySize]byte)
	for i := range validShard {
		validShard[i] = byte(i)
	}
	scr.Shard = validShard

	payload, marshalErr = json.Marshal(scr)
	if marshalErr != nil {
		t.Fatalf("Failed to marshal ShardPutRequest with shard: %v", marshalErr)
	}

	unmarshalErr = json.Unmarshal(payload, &unmarshaled)
	if unmarshalErr != nil {
		t.Fatalf("Failed to unmarshal payload with shard: %v", unmarshalErr)
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
	// Test HTTP client behavior for shard contribution requests
	testPayload := []byte(`{"shard": "test data"}`)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify the request structure
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST method, got %s", r.Method)
		}

		body, readErr := io.ReadAll(r.Body)
		if readErr != nil {
			t.Errorf("Failed to read request body: %v", readErr)
			return
		}

		if !bytes.Equal(body, testPayload) {
			t.Errorf("Request body mismatch. Expected: %s, Got: %s",
				string(testPayload), string(body))
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	}))
	defer server.Close()

	// Test a successful HTTP POST request
	client := server.Client()
	req, reqErr := http.NewRequest(
		http.MethodPost, server.URL, bytes.NewReader(testPayload),
	)
	if reqErr != nil {
		t.Fatalf("Failed to create request: %v", reqErr)
	}

	resp, doErr := client.Do(req)
	if doErr != nil {
		t.Fatalf("Failed to send request: %v", doErr)
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	respBody, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		t.Fatalf("Failed to read response: %v", readErr)
	}

	if string(respBody) != "OK" {
		t.Errorf("Expected response 'OK', got '%s'", string(respBody))
	}
}
