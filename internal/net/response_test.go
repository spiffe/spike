//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package net

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

type testResponse struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
}

func TestMarshalBodyAndRespondOnMarshalFail_Success(t *testing.T) {
	w := httptest.NewRecorder()
	res := testResponse{Message: "success", Code: 200}

	body, err := MarshalBodyAndRespondOnMarshalFail(res, w)

	if err != nil {
		t.Errorf("MarshalBodyAndRespondOnMarshalFail() error = %v, want nil", err)
	}

	if body == nil {
		t.Error("MarshalBodyAndRespondOnMarshalFail() body is nil")
	}

	// Verify JSON is valid
	var decoded testResponse
	if unmarshalErr := json.Unmarshal(body, &decoded); unmarshalErr != nil {
		t.Errorf("MarshalBodyAndRespondOnMarshalFail() invalid JSON: %v",
			unmarshalErr)
	}

	if decoded.Message != "success" {
		t.Errorf("MarshalBodyAndRespondOnMarshalFail() message = %q, want %q",
			decoded.Message, "success")
	}

	// No response should be written on success
	if w.Code != http.StatusOK {
		t.Errorf("MarshalBodyAndRespondOnMarshalFail() wrote status %d on success",
			w.Code)
	}
}

func TestMarshalBodyAndRespondOnMarshalFail_UnmarshalableType(t *testing.T) {
	w := httptest.NewRecorder()
	// Channels cannot be marshaled to JSON
	res := make(chan int)

	body, err := MarshalBodyAndRespondOnMarshalFail(res, w)

	if err == nil {
		t.Error("MarshalBodyAndRespondOnMarshalFail() expected error for channel")
	}

	if body != nil {
		t.Error("MarshalBodyAndRespondOnMarshalFail() body should be nil on error")
	}

	if w.Code != http.StatusInternalServerError {
		t.Errorf("MarshalBodyAndRespondOnMarshalFail() status = %d, want %d",
			w.Code, http.StatusInternalServerError)
	}
}

func TestRespond_SetsHeaders(t *testing.T) {
	w := httptest.NewRecorder()
	body := []byte(`{"test":"value"}`)

	Respond(http.StatusOK, body, w)

	// Check Content-Type
	if ct := w.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Respond() Content-Type = %q, want %q", ct, "application/json")
	}

	// Check Cache-Control
	cc := w.Header().Get("Cache-Control")
	if cc != "no-store, no-cache, must-revalidate, private" {
		t.Errorf("Respond() Cache-Control = %q, want no-cache headers", cc)
	}

	// Check Pragma
	if pragma := w.Header().Get("Pragma"); pragma != "no-cache" {
		t.Errorf("Respond() Pragma = %q, want %q", pragma, "no-cache")
	}

	// Check Expires
	if expires := w.Header().Get("Expires"); expires != "0" {
		t.Errorf("Respond() Expires = %q, want %q", expires, "0")
	}

	// Check status code
	if w.Code != http.StatusOK {
		t.Errorf("Respond() status = %d, want %d", w.Code, http.StatusOK)
	}

	// Check body
	if w.Body.String() != `{"test":"value"}` {
		t.Errorf("Respond() body = %q, want %q",
			w.Body.String(), `{"test":"value"}`)
	}
}

func TestRespond_DifferentStatusCodes(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
	}{
		{"OK", http.StatusOK},
		{"Created", http.StatusCreated},
		{"BadRequest", http.StatusBadRequest},
		{"Unauthorized", http.StatusUnauthorized},
		{"Forbidden", http.StatusForbidden},
		{"NotFound", http.StatusNotFound},
		{"InternalServerError", http.StatusInternalServerError},
		{"ServiceUnavailable", http.StatusServiceUnavailable},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			body := []byte(`{}`)

			Respond(tt.statusCode, body, w)

			if w.Code != tt.statusCode {
				t.Errorf("Respond() status = %d, want %d", w.Code, tt.statusCode)
			}
		})
	}
}
