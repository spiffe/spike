//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package reqres

import (
	"github.com/spiffe/spike/internal/entity/data"
	"time"
)

// RootKeyCacheRequest is to cache the generated root key in SPIKE Keep.
// If the root key is lost due to a crash, it will be retrieved from SPIKE Keep.
type RootKeyCacheRequest struct {
	RootKey string `json:"rootKey"`
}

type ErrorCode string

var ErrBadInput = ErrorCode("bad_request")
var ErrServerFault = ErrorCode("server_fault")
var ErrUnauthorized = ErrorCode("unauthorized")
var ErrLowEntropy = ErrorCode("low_entropy")
var ErrAlreadyInitialized = ErrorCode("already_initialized")
var ErrNotFound = ErrorCode("not_found")

type FallbackResponse struct {
	Err ErrorCode `json:"err,omitempty"`
}

// RootKeyCacheResponse is to cache the generated root key in SPIKE Keep.
type RootKeyCacheResponse struct {
	Err ErrorCode `json:"error,omitempty"`
}

// RootKeyReadRequest is a request to get the root key back from remote cache.
type RootKeyReadRequest struct{}

// RootKeyReadResponse is a response to get the root key back from remote cache.
type RootKeyReadResponse struct {
	RootKey string    `json:"rootKey"`
	Err     ErrorCode `json:"err,omitempty"`
}

// AdminTokenWriteRequest is to persist the admin token in memory.
// Admin token can be persisted only once. It is used to receive a
// short-lived session token.
type AdminTokenWriteRequest struct {
	Data string `json:"data"`
}

// AdminTokenWriteResponse is to persist the admin token in memory.
type AdminTokenWriteResponse struct {
	Err ErrorCode `json:"err,omitempty"`
}

type CheckInitStateRequest struct {
}

type CheckInitStateResponse struct {
	State data.InitState `json:"state"`
	Err   ErrorCode      `json:"err,omitempty"`
}

type InitRequest struct {
	Password string `json:"password"`
}

type InitResponse struct {
	Err ErrorCode `json:"err,omitempty"`
}

type AdminLoginRequest struct {
	Password string `json:"password"`
}

type AdminLoginResponse struct {
	Token     string    `json:"token"`
	Signature string    `json:"signature"`
	Err       ErrorCode `json:"err,omitempty"`
}

// SecretResponseMetadata is meta information about secrets for internal tracking.
type SecretResponseMetadata struct {
	CreatedTime time.Time  `json:"created_time"`
	Version     int        `json:"version"`
	DeletedTime *time.Time `json:"deleted_time,omitempty"`
}

// SecretPutRequest for creating/updating secrets
type SecretPutRequest struct {
	Path   string            `json:"path"`
	Values map[string]string `json:"values"`
	Err    ErrorCode         `json:"err,omitempty"`
}

// SecretPutResponse is after successful secret write
type SecretPutResponse struct {
	SecretResponseMetadata
	Err ErrorCode `json:"err,omitempty"`
}

// SecretReadRequest is for getting secrets
type SecretReadRequest struct {
	Path    string `json:"path"`
	Version int    `json:"version,omitempty"` // Optional specific version
}

// SecretReadResponse is for getting secrets
type SecretReadResponse struct {
	Data map[string]string `json:"data"`
	Err  ErrorCode         `json:"err,omitempty"`
}

// SecretDeleteRequest for soft-deleting secret versions
type SecretDeleteRequest struct {
	Path     string `json:"path"`
	Versions []int  `json:"versions"` // Empty means latest version
}

// SecretDeleteResponse after soft-delete
type SecretDeleteResponse struct {
	Metadata SecretResponseMetadata `json:"metadata"`
	Err      ErrorCode              `json:"err,omitempty"`
}

// SecretUndeleteRequest for recovering soft-deleted versions
type SecretUndeleteRequest struct {
	Path     string `json:"path"`
	Versions []int  `json:"versions"`
}

// SecretUndeleteResponse after recovery
type SecretUndeleteResponse struct {
	Metadata SecretResponseMetadata `json:"metadata"`
	Err      ErrorCode              `json:"err,omitempty"`
}

// SecretListRequest for listing secrets
type SecretListRequest struct {
}

// SecretListResponse for listing secrets
type SecretListResponse struct {
	Keys []string  `json:"keys"`
	Err  ErrorCode `json:"err,omitempty"`
}
