//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package net

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/spiffe/spike-sdk-go/net"
)

type testRequest struct {
	Name  string `json:"name"`
	Value int    `json:"value"`
}

type testErrorResponse struct {
	Err string `json:"err"`
}

func TestUnmarshalAndRespondOnFail_Success(t *testing.T) {
	w := httptest.NewRecorder()
	requestBody := []byte(`{"name":"test","value":42}`)
	errorResp := testErrorResponse{Err: "bad request"}

	result, err := net.UnmarshalAndRespondOnFail[testRequest, testErrorResponse](
		requestBody, w, errorResp,
	)

	if err != nil {
		t.Fatalf("UnmarshalAndRespondOnFail() error = %v, want nil", err)
	}

	if result == nil {
		t.Fatal("UnmarshalAndRespondOnFail() result is nil")
		return
	}

	if result.Name != "test" {
		t.Errorf("UnmarshalAndRespondOnFail() Name = %q, want %q",
			result.Name, "test")
	}

	if result.Value != 42 {
		t.Errorf("UnmarshalAndRespondOnFail() Value = %d, want %d",
			result.Value, 42)
	}

	// No response should be written on success
	if w.Body.Len() > 0 {
		t.Errorf("UnmarshalAndRespondOnFail() wrote body on success: %q",
			w.Body.String())
	}
}

func TestUnmarshalAndRespondOnFail_InvalidJSON(t *testing.T) {
	w := httptest.NewRecorder()
	requestBody := []byte(`{invalid json}`)
	errorResp := testErrorResponse{Err: "bad request"}

	result, err := net.UnmarshalAndRespondOnFail[testRequest, testErrorResponse](
		requestBody, w, errorResp,
	)

	if err == nil {
		t.Error("UnmarshalAndRespondOnFail() expected error for invalid JSON")
	}

	if result != nil {
		t.Errorf("UnmarshalAndRespondOnFail() result = %v, want nil", result)
	}

	if w.Code != http.StatusBadRequest {
		t.Errorf("UnmarshalAndRespondOnFail() status = %d, want %d",
			w.Code, http.StatusBadRequest)
	}
}

func TestUnmarshalAndRespondOnFail_EmptyBody(t *testing.T) {
	w := httptest.NewRecorder()
	requestBody := []byte(``)
	errorResp := testErrorResponse{Err: "bad request"}

	result, err := net.UnmarshalAndRespondOnFail[testRequest, testErrorResponse](
		requestBody, w, errorResp,
	)

	if err == nil {
		t.Error("UnmarshalAndRespondOnFail() expected error for empty body")
	}

	if result != nil {
		t.Errorf("UnmarshalAndRespondOnFail() result = %v, want nil", result)
	}
}

func TestFail_SendsErrorResponse(t *testing.T) {
	w := httptest.NewRecorder()
	errorResp := testErrorResponse{Err: "something went wrong"}

	_ = net.Fail(errorResp, w, http.StatusBadRequest)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Fail() status = %d, want %d", w.Code, http.StatusBadRequest)
	}

	// Verify JSON response
	var decoded testErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &decoded); err != nil {
		t.Errorf("Fail() invalid JSON response: %v", err)
	}

	if decoded.Err != "something went wrong" {
		t.Errorf("Fail() Err = %q, want %q", decoded.Err, "something went wrong")
	}
}

func TestSuccess_SendsOKResponse(t *testing.T) {
	w := httptest.NewRecorder()
	resp := testResponse{Message: "created", Code: 200}

	_ = net.Success(resp, w)

	if w.Code != http.StatusOK {
		t.Errorf("Success() status = %d, want %d", w.Code, http.StatusOK)
	}

	// Verify JSON response
	var decoded testResponse
	if err := json.Unmarshal(w.Body.Bytes(), &decoded); err != nil {
		t.Errorf("Success() invalid JSON response: %v", err)
	}

	if decoded.Message != "created" {
		t.Errorf("Success() Message = %q, want %q", decoded.Message, "created")
	}
}

func TestSuccessWithResponseBody_ReturnsBody(t *testing.T) {
	w := httptest.NewRecorder()
	resp := testResponse{Message: "data", Code: 200}

	body, _ := net.SuccessWithResponseBody(resp, w)

	if body == nil {
		t.Error("SuccessWithResponseBody() returned nil body")
	}

	if w.Code != http.StatusOK {
		t.Errorf("SuccessWithResponseBody() status = %d, want %d",
			w.Code, http.StatusOK)
	}

	// Verify body matches response
	var decoded testResponse
	if err := json.Unmarshal(body, &decoded); err != nil {
		t.Errorf("SuccessWithResponseBody() invalid JSON body: %v", err)
	}

	if decoded.Message != "data" {
		t.Errorf("SuccessWithResponseBody() Message = %q, want %q",
			decoded.Message, "data")
	}
}

func TestReadRequestBodyAndRespondOnFail_Success(t *testing.T) {
	expectedBody := `{"test":"data"}`
	r := httptest.NewRequest(http.MethodPost, "/test",
		bytes.NewBufferString(expectedBody))
	w := httptest.NewRecorder()

	body, err := net.ReadRequestBodyAndRespondOnFail(w, r)

	if err != nil {
		t.Errorf("ReadRequestBodyAndRespondOnFail() error = %v, want nil", err)
	}

	if string(body) != expectedBody {
		t.Errorf("ReadRequestBodyAndRespondOnFail() body = %q, want %q",
			string(body), expectedBody)
	}
}

func TestFail_DifferentStatusCodes(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
	}{
		{"BadRequest", http.StatusBadRequest},
		{"Unauthorized", http.StatusUnauthorized},
		{"Forbidden", http.StatusForbidden},
		{"NotFound", http.StatusNotFound},
		{"InternalServerError", http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			errorResp := testErrorResponse{Err: "error"}

			_ = net.Fail(errorResp, w, tt.statusCode)

			if w.Code != tt.statusCode {
				t.Errorf("Fail() status = %d, want %d", w.Code, tt.statusCode)
			}
		})
	}
}
