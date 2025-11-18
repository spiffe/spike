//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package errors

// NotReadyError checks if an error indicates SPIKE Nexus is not ready.
//
// This function uses string comparison because the error originates from
// a JSON response (internal/net/response.go:165 returns data.ErrNotReady),
// gets serialized over the wire, and is then wrapped by the SDK via
// errors.Join(), which breaks errors.Is() checking.
//
// The SDK (spike-sdk-go/.../decrypt.go:142) casts res.Err (ErrorCode) to
// the string "not ready". Since this error code is unique to the not-ready
// condition and no other errors use this exact string, exact string matching
// is safe (though not ideal).
//
// Parameters:
//   - err: The error to check
//
// Returns:
//   - true if the error indicates the server is not ready
//   - false otherwise (including if err is nil)
//
// Technical debt: This should be replaced with proper error codes or typed
// errors if the SDK adds support for that in the future. The current approach
// is necessary because sentinel errors do not survive JSON serialization and
// error wrapping.
func NotReadyError(err error) bool {
	return err != nil && err.Error() == "not ready"
}
