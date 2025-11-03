//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package kek

import (
	"net/http"

	"github.com/spiffe/spike-sdk-go/api/entity/data"
	"github.com/spiffe/spike-sdk-go/api/errors"
	"github.com/spiffe/spike-sdk-go/log"

	"github.com/spiffe/spike/app/nexus/internal/state/persist"
	"github.com/spiffe/spike/internal/journal"
	"github.com/spiffe/spike/internal/net"
)

// RouteRotateKEK handles manual KEK rotation requests
//
// This endpoint allows administrators to manually trigger a KEK rotation
// outside of the automatic schedule. The rotation will:
// 1. Create a new KEK with a new version
// 2. Move the old KEK to grace period
// 3. Start using the new KEK for new secret wraps
//
// Request: POST /v1/kek/rotate
// Response: { "success": true, "new_kek_id": "v2-2025-01", "message": "..." }
func RouteRotateKEK(
	w http.ResponseWriter, r *http.Request, audit *journal.AuditEntry,
) error {
	const fName = "routeRotateKEK"
	journal.AuditRequest(fName, r, audit, journal.AuditCreate)

	// Check if KEK manager is initialized
	if !persist.IsKEKManagerInitialized() {
		log.Log().Warn(fName, "message", "KEK manager not initialized")
		responseBody := net.MarshalBody(map[string]any{
			"err":     data.ErrBadInput,
			"message": "KEK rotation not enabled",
		}, w)
		if responseBody == nil {
			return errors.ErrMarshalFailure
		}
		net.Respond(http.StatusBadRequest, responseBody, w)
		return nil
	}

	manager := persist.GetKEKManager()
	if manager == nil {
		responseBody := net.MarshalBody(map[string]any{
			"err":     data.ErrBadInput,
			"message": "KEK manager unavailable",
		}, w)
		if responseBody == nil {
			return errors.ErrMarshalFailure
		}
		net.Respond(http.StatusInternalServerError, responseBody, w)
		return nil
	}

	// Perform rotation
	err := manager.RotateKEK()
	if err != nil {
		log.Log().Error(fName, "message", "KEK rotation failed", "err", err.Error())
		responseBody := net.MarshalBody(map[string]any{
			"err":     data.ErrBadInput,
			"message": "KEK rotation failed: " + err.Error(),
		}, w)
		if responseBody == nil {
			return errors.ErrMarshalFailure
		}
		net.Respond(http.StatusInternalServerError, responseBody, w)
		return nil
	}

	newKekID := manager.GetCurrentKEKID()

	log.Log().Info(fName, "message", "KEK rotation completed", "new_kek_id", newKekID)

	responseBody := net.MarshalBody(map[string]any{
		"success":    true,
		"new_kek_id": newKekID,
		"message":    "KEK rotation completed successfully",
	}, w)
	if responseBody == nil {
		return errors.ErrMarshalFailure
	}

	net.Respond(http.StatusOK, responseBody, w)
	return nil
}
