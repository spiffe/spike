//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package net

import (
	"net/http"

	"github.com/spiffe/spike-sdk-go/api/entity/v1/reqres"
	sdkErrors "github.com/spiffe/spike-sdk-go/errors"
)

func respondFallbackWithStatus(
	w http.ResponseWriter, status int, code sdkErrors.ErrorCode,
) *sdkErrors.SDKError {
	body, err := MarshalBodyAndRespondOnMarshalFail(
		reqres.FallbackResponse{Err: code}, w,
	)
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json")

	// Add cache invalidation headers
	w.Header().Set(
		"Cache-Control",
		"no-store, no-cache, must-revalidate, private",
	)
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	w.WriteHeader(status)

	if _, err := w.Write(body); err != nil {
		failErr := sdkErrors.ErrAPIInternal.Wrap(err)
		return failErr
	}

	return nil
}
