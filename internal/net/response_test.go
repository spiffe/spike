//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package net

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/spiffe/spike-sdk-go/net"

	sdkErrors "github.com/spiffe/spike-sdk-go/errors"
)

type testResponse struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
}

func TestMarshalBodyAndRespondOnMarshalFail_Success(t *testing.T) {
	w := httptest.NewRecorder()
	res := testResponse{Message: "success", Code: 200}

	body, err := net.MarshalBodyAndRespondOnMarshalFail(res, w)

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

	body, err := net.MarshalBodyAndRespondOnMarshalFail(res, w)

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

	_ = net.Respond(http.StatusOK, body, w)

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

			_ = net.Respond(tt.statusCode, body, w)

			if w.Code != tt.statusCode {
				t.Errorf("Respond() status = %d, want %d", w.Code, tt.statusCode)
			}
		})
	}
}

// mockErrorResponse implements ErrorResponder for testing HandleError.
type mockErrorResponse struct {
	Err string `json:"err"`
}

func (m mockErrorResponse) NotFound() mockErrorResponse {
	return mockErrorResponse{Err: "not_found"}
}

func (m mockErrorResponse) Internal() mockErrorResponse {
	return mockErrorResponse{Err: "internal"}
}

func TestHandleError_NilError(t *testing.T) {
	w := httptest.NewRecorder()

	result := net.RespondWithHTTPError(nil, w, mockErrorResponse{})

	if result != nil {
		t.Errorf("HandleError(nil) = %v, want nil", result)
	}

	// Response should not have been written
	if w.Code != http.StatusOK {
		t.Errorf("HandleError(nil) wrote response, status = %d", w.Code)
	}
}

func TestHandleError_NotFoundError(t *testing.T) {
	w := httptest.NewRecorder()

	err := sdkErrors.ErrEntityNotFound.Clone()
	result := net.RespondWithHTTPError(err, w, mockErrorResponse{})

	if result == nil {
		t.Error("HandleError() returned nil for not found error")
	}

	if w.Code != http.StatusNotFound {
		t.Errorf("HandleError() status = %d, want %d",
			w.Code, http.StatusNotFound)
	}
}

func TestHandleError_OtherError(t *testing.T) {
	w := httptest.NewRecorder()

	err := sdkErrors.ErrAPIBadRequest.Clone()
	result := net.RespondWithHTTPError(err, w, mockErrorResponse{})

	if result == nil {
		t.Error("HandleError() returned nil for other error")
	}

	if w.Code != http.StatusInternalServerError {
		t.Errorf("HandleError() status = %d, want %d",
			w.Code, http.StatusInternalServerError)
	}
}

func TestHandleError_WrappedNotFoundError(t *testing.T) {
	w := httptest.NewRecorder()

	// Create a wrapped not found error
	wrappedErr := sdkErrors.ErrEntityNotFound.Wrap(sdkErrors.ErrAPIBadRequest)
	result := net.RespondWithHTTPError(wrappedErr, w, mockErrorResponse{})

	if result == nil {
		t.Error("HandleError() returned nil for wrapped not found error")
	}

	// Should still be recognized as not found
	if w.Code != http.StatusNotFound {
		t.Errorf("HandleError() status = %d, want %d",
			w.Code, http.StatusNotFound)
	}
}

func TestHandleInternalError(t *testing.T) {
	w := httptest.NewRecorder()

	err := sdkErrors.ErrCryptoCipherNotAvailable.Clone()
	result := net.RespondWithInternalError(err, w, mockErrorResponse{})

	if result == nil {
		t.Error("HandleInternalError() returned nil")
	}

	if !errors.Is(result, err) {
		t.Error("HandleInternalError() should return the same error")
	}

	if w.Code != http.StatusInternalServerError {
		t.Errorf("HandleInternalError() status = %d, want %d",
			w.Code, http.StatusInternalServerError)
	}
}
