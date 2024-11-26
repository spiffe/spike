//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package reqres

// RootKeyCacheRequest is to cache the generated root key in SPIKE Keep.
// If the root key is lost due to a crash, it will be retrieved from SPIKE Keep.
type RootKeyCacheRequest struct {
	RootKey string `json:"rootKey"`
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
