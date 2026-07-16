//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package operator

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/spiffe/spike-sdk-go/config/env"
	sdkErrors "github.com/spiffe/spike-sdk-go/errors"
	"github.com/spiffe/spike-sdk-go/journal"
)

func TestRouteRecover_MemoryMode(t *testing.T) {
	// Save original environment variables
	originalStore := os.Getenv(env.NexusBackendStore)
	defer func() {
		if originalStore != "" {
			_ = os.Setenv(env.NexusBackendStore, originalStore)
		} else {
			_ = os.Unsetenv(env.NexusBackendStore)
		}
	}()

	// Set to memory mode
	_ = os.Setenv(env.NexusBackendStore, "memory")

	// Verify the environment is set correctly
	if env.BackendStoreTypeVal() != env.Memory {
		t.Fatal("Expected Memory backend store type")
	}

	// Create a test request
	req := httptest.NewRequest(http.MethodPost, "/recover", nil)
	w := httptest.NewRecorder()
	audit := &journal.AuditEntry{}

	// Call function
	err := RouteRecover(w, req, audit)

	// Recovery shards are useless for the in-memory backend; the route
	// must reject the request explicitly rather than fail later with a
	// misleading "not enough shards" internal error.
	if err == nil {
		t.Error("Expected an error in memory mode")
		return
	}
	if !err.Is(sdkErrors.ErrDataInvalidInput) {
		t.Errorf("Expected ErrDataInvalidInput, got: %v", err)
	}
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got: %d",
			http.StatusBadRequest, w.Code)
	}
}
