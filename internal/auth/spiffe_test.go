//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package auth

import (
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

// testErrorResponse is a simple struct for testing error responses
type testErrorResponse struct {
	Err string `json:"err"`
}

// createMockRequest creates an HTTP request with a mock TLS connection
// containing the specified SPIFFE ID in the peer certificate's URI SAN.
func createMockRequest(spiffeID string) *http.Request {
	req := httptest.NewRequest(http.MethodGet, "/test", nil)

	if spiffeID == "" {
		// No TLS connection
		return req
	}

	// Parse the SPIFFE ID as a URL
	spiffeURL, err := url.Parse(spiffeID)
	if err != nil {
		return req
	}

	// Create a mock certificate with the SPIFFE ID as a URI SAN
	cert := &x509.Certificate{
		Subject: pkix.Name{
			CommonName: "test-workload",
		},
		URIs: []*url.URL{spiffeURL},
	}

	// Create mock TLS connection state
	req.TLS = &tls.ConnectionState{
		PeerCertificates: []*x509.Certificate{cert},
	}

	return req
}

func TestExtractPeerSPIFFEID_NoPeerCertificates(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.TLS = &tls.ConnectionState{
		PeerCertificates: []*x509.Certificate{},
	}
	w := httptest.NewRecorder()

	errorResponse := testErrorResponse{Err: "unauthorized"}

	peerID, err := ExtractPeerSPIFFEID(req, w, errorResponse)

	if peerID != nil {
		t.Errorf("ExtractPeerSPIFFEID() peerID = %v, want nil", peerID)
	}
	if err == nil {
		t.Error("ExtractPeerSPIFFEID() expected error, got nil")
	}

	if w.Code != http.StatusUnauthorized {
		t.Errorf("ExtractPeerSPIFFEID() status = %d, want %d",
			w.Code, http.StatusUnauthorized)
	}
}

func TestExtractPeerSPIFFEID_ValidSPIFFEID(t *testing.T) {
	validSPIFFEID := "spiffe://example.org/workload/test"
	req := createMockRequest(validSPIFFEID)
	w := httptest.NewRecorder()

	errorResponse := testErrorResponse{Err: "unauthorized"}

	peerID, err := ExtractPeerSPIFFEID(req, w, errorResponse)

	if err != nil {
		t.Errorf("ExtractPeerSPIFFEID() unexpected error: %v", err)
		return
	}

	if peerID == nil {
		t.Error("ExtractPeerSPIFFEID() peerID is nil, want valid ID")
		return
	}

	if peerID.String() != validSPIFFEID {
		t.Errorf("ExtractPeerSPIFFEID() peerID = %q, want %q",
			peerID.String(), validSPIFFEID)
	}

	// Response should not have been written for success case
	if w.Code != http.StatusOK {
		t.Errorf("ExtractPeerSPIFFEID() wrote response on success, status = %d",
			w.Code)
	}
}

func TestExtractPeerSPIFFEID_InvalidSPIFFEID(t *testing.T) {
	// Invalid SPIFFE ID (not a valid URI scheme)
	invalidSPIFFEID := "http://example.org/workload/test"
	req := createMockRequest(invalidSPIFFEID)
	w := httptest.NewRecorder()

	errorResponse := testErrorResponse{Err: "unauthorized"}

	peerID, err := ExtractPeerSPIFFEID(req, w, errorResponse)

	if peerID != nil {
		t.Errorf("ExtractPeerSPIFFEID() peerID = %v, want nil", peerID)
	}
	if err == nil {
		t.Error("ExtractPeerSPIFFEID() expected error for invalid SPIFFE ID")
	}

	if w.Code != http.StatusUnauthorized {
		t.Errorf("ExtractPeerSPIFFEID() status = %d, want %d",
			w.Code, http.StatusUnauthorized)
	}
}

func TestExtractPeerSPIFFEID_CertWithNoURIs(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test", nil)

	// Create a certificate with no URIs
	cert := &x509.Certificate{
		Subject: pkix.Name{
			CommonName: "test-workload",
		},
		URIs: nil, // No URIs
	}

	req.TLS = &tls.ConnectionState{
		PeerCertificates: []*x509.Certificate{cert},
	}

	w := httptest.NewRecorder()
	errorResponse := testErrorResponse{Err: "unauthorized"}

	peerID, err := ExtractPeerSPIFFEID(req, w, errorResponse)

	if peerID != nil {
		t.Errorf("ExtractPeerSPIFFEID() peerID = %v, want nil", peerID)
	}
	if err == nil {
		t.Error("ExtractPeerSPIFFEID() expected error for cert with no URIs")
	}

	if w.Code != http.StatusUnauthorized {
		t.Errorf("ExtractPeerSPIFFEID() status = %d, want %d",
			w.Code, http.StatusUnauthorized)
	}
}

func TestExtractPeerSPIFFEID_MultipleSPIFFEIDs(t *testing.T) {
	tests := []struct {
		name      string
		spiffeID  string
		wantValid bool
	}{
		{
			name:      "standard workload ID",
			spiffeID:  "spiffe://example.org/workload/app",
			wantValid: true,
		},
		{
			name:      "ID with multiple path segments",
			spiffeID:  "spiffe://example.org/region/us-west/service/api",
			wantValid: true,
		},
		{
			name:      "ID with simple path",
			spiffeID:  "spiffe://example.org/workload",
			wantValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := createMockRequest(tt.spiffeID)
			w := httptest.NewRecorder()
			errorResponse := testErrorResponse{Err: "unauthorized"}

			peerID, err := ExtractPeerSPIFFEID(req, w, errorResponse)

			if tt.wantValid {
				if err != nil {
					t.Errorf("ExtractPeerSPIFFEID() unexpected error: %v", err)
					return
				}
				if peerID == nil {
					t.Error("ExtractPeerSPIFFEID() peerID is nil")
					return
				}
				if peerID.String() != tt.spiffeID {
					t.Errorf("ExtractPeerSPIFFEID() = %q, want %q",
						peerID.String(), tt.spiffeID)
				}
			} else {
				if err == nil {
					t.Error("ExtractPeerSPIFFEID() expected error")
				}
			}
		})
	}
}

func TestExtractPeerSPIFFEID_DifferentErrorResponses(t *testing.T) {
	// Test that the function works with different error response types
	// Use a request with TLS but no peer certificates to trigger error
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.TLS = &tls.ConnectionState{
		PeerCertificates: []*x509.Certificate{},
	}

	t.Run("with string error response", func(t *testing.T) {
		w := httptest.NewRecorder()
		_, err := ExtractPeerSPIFFEID(req, w, "error string")
		if err == nil {
			t.Error("Expected error")
		}
	})

	t.Run("with struct error response", func(t *testing.T) {
		w := httptest.NewRecorder()
		_, err := ExtractPeerSPIFFEID(req, w, testErrorResponse{Err: "test"})
		if err == nil {
			t.Error("Expected error")
		}
	})

	t.Run("with map error response", func(t *testing.T) {
		w := httptest.NewRecorder()
		_, err := ExtractPeerSPIFFEID(req, w, map[string]string{"err": "test"})
		if err == nil {
			t.Error("Expected error")
		}
	})
}
