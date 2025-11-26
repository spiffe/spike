//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package secret

import (
	"net/http"
	"net/http/httptest"
	"testing"

	sdkErrors "github.com/spiffe/spike-sdk-go/errors"
)

func TestHandleGetSecretError_NilError(t *testing.T) {
	w := httptest.NewRecorder()

	result := handleGetSecretError(nil, w)

	if result != nil {
		t.Errorf("handleGetSecretError(nil) = %v, want nil", result)
	}

	// Response should not have been written
	if w.Code != http.StatusOK {
		t.Errorf("handleGetSecretError(nil) wrote response, status = %d",
			w.Code)
	}
}

func TestHandleGetSecretError_NotFoundError(t *testing.T) {
	w := httptest.NewRecorder()

	err := sdkErrors.ErrEntityNotFound
	result := handleGetSecretError(err, w)

	if result == nil {
		t.Error("handleGetSecretError() returned nil for not found error")
	}

	if w.Code != http.StatusNotFound {
		t.Errorf("handleGetSecretError() status = %d, want %d",
			w.Code, http.StatusNotFound)
	}
}

func TestHandleGetSecretError_OtherError(t *testing.T) {
	w := httptest.NewRecorder()

	err := sdkErrors.ErrAPIBadRequest
	result := handleGetSecretError(err, w)

	if result == nil {
		t.Error("handleGetSecretError() returned nil for other error")
	}

	if w.Code != http.StatusInternalServerError {
		t.Errorf("handleGetSecretError() status = %d, want %d",
			w.Code, http.StatusInternalServerError)
	}
}

func TestHandleGetSecretMetadataError_NilError(t *testing.T) {
	w := httptest.NewRecorder()

	result := handleGetSecretMetadataError(nil, w)

	if result != nil {
		t.Errorf("handleGetSecretMetadataError(nil) = %v, want nil", result)
	}

	// Response should not have been written
	if w.Code != http.StatusOK {
		t.Errorf("handleGetSecretMetadataError(nil) wrote response, status = %d",
			w.Code)
	}
}

func TestHandleGetSecretMetadataError_NotFoundError(t *testing.T) {
	w := httptest.NewRecorder()

	err := sdkErrors.ErrEntityNotFound
	result := handleGetSecretMetadataError(err, w)

	if result == nil {
		t.Error("handleGetSecretMetadataError() returned nil for not found error")
	}

	if w.Code != http.StatusNotFound {
		t.Errorf("handleGetSecretMetadataError() status = %d, want %d",
			w.Code, http.StatusNotFound)
	}
}

func TestHandleGetSecretMetadataError_OtherError(t *testing.T) {
	w := httptest.NewRecorder()

	err := sdkErrors.ErrAPIBadRequest
	result := handleGetSecretMetadataError(err, w)

	if result == nil {
		t.Error("handleGetSecretMetadataError() returned nil for other error")
	}

	if w.Code != http.StatusInternalServerError {
		t.Errorf("handleGetSecretMetadataError() status = %d, want %d",
			w.Code, http.StatusInternalServerError)
	}
}

func TestHandleGetSecretError_WrappedNotFoundError(t *testing.T) {
	w := httptest.NewRecorder()

	// Create a wrapped not found error
	wrappedErr := sdkErrors.ErrEntityNotFound.Wrap(
		sdkErrors.ErrAPIBadRequest,
	)
	result := handleGetSecretError(wrappedErr, w)

	if result == nil {
		t.Error("handleGetSecretError() returned nil for wrapped not found error")
	}

	// Should still be recognized as not found
	if w.Code != http.StatusNotFound {
		t.Errorf("handleGetSecretError() status = %d, want %d",
			w.Code, http.StatusNotFound)
	}
}

func TestHandleGetSecretMetadataError_WrappedNotFoundError(t *testing.T) {
	w := httptest.NewRecorder()

	// Create a wrapped not found error
	wrappedErr := sdkErrors.ErrEntityNotFound.Wrap(
		sdkErrors.ErrAPIBadRequest,
	)
	result := handleGetSecretMetadataError(wrappedErr, w)

	if result == nil {
		t.Error(
			"handleGetSecretMetadataError() returned nil for wrapped not found",
		)
	}

	// Should still be recognized as not found
	if w.Code != http.StatusNotFound {
		t.Errorf("handleGetSecretMetadataError() status = %d, want %d",
			w.Code, http.StatusNotFound)
	}
}
