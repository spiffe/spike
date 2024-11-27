//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package reqres

import (
	"time"
)

// SecretResponseMetadata is meta information about secrets for internal
// tracking.
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

// SecretMetadataReadRequest for get secrets metadata
type SecretMetadataReadRequest struct {
	Path    string `json:"path"`
	Version int    `json:"version,omitempty"` // Optional specific version
}

// SecretMetadataResponse for secrets versions and metadata
type SecretMetadataResponse struct {
	Versions map[int]SecretMetadataVersionResponse `json:"versions,omitempty"`
	Metadata SecretRawMetadataResponse             `json:"metadata,omitempty"`
	Err      ErrorCode                             `json:"err,omitempty"`
}

// SecretMetadataVersionResponse for secrets version
type SecretMetadataVersionResponse struct {
	CreatedTime time.Time  `json:"createdTime"`
	Version     int        `json:"version"`
	DeletedTime *time.Time `json:"deletedTime"`
}

// SecretRawMetadataResponse for secrets raw metadata
type SecretRawMetadataResponse struct {
	CurrentVersion int       `json:"currentVersion"`
	OldestVersion  int       `json:"oldestVersion"`
	CreatedTime    time.Time `json:"createdTime"`
	UpdatedTime    time.Time `json:"updatedTime"`
	MaxVersions    int       `json:"maxVersions"`
}
