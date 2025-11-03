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

// RouteRotateRMK handles RMK (Root Master Key) rotation ceremony
//
// This is a critical operation that should only be performed during maintenance.
// It rewraps all KEKs with a new RMK without touching secret data.
//
// Request: POST /v1/rmk/rotate
// Body: { "new_rmk_version": 2 }
// Response: { "success": true, "keks_rewrapped": 5, "duration_ms": 123 }
//
// NOTE: This is a placeholder implementation. In production, this would:
// 1. Require special admin credentials
// 2. Put system in maintenance mode
// 3. Perform M-of-N ceremony for new RMK
// 4. Validate the rotation
// 5. Return detailed results
func RouteRotateRMK(
	w http.ResponseWriter, r *http.Request, audit *journal.AuditEntry,
) error {
	const fName = "routeRotateRMK"
	journal.AuditRequest(fName, r, audit, journal.AuditCreate)

	log.Log().Warn(fName, "message", "RMK rotation endpoint called - NOT IMPLEMENTED")

	responseBody := net.MarshalBody(map[string]any{
		"err":           data.ErrBadInput,
		"message":       "RMK rotation requires manual ceremony - use CLI tools",
		"documentation": "See docs for RMK rotation procedure",
	}, w)
	if responseBody == nil {
		return errors.ErrMarshalFailure
	}

	net.Respond(http.StatusNotImplemented, responseBody, w)
	return nil
}

// RouteRMKSnapshot creates a snapshot of current KEK state for RMK rotation
//
// Request: GET /v1/rmk/snapshot
// Response: { "rmk_version": 1, "kek_count": 5, "snapshot_id": "..." }
func RouteRMKSnapshot(
	w http.ResponseWriter, r *http.Request, audit *journal.AuditEntry,
) error {
	const fName = "routeRMKSnapshot"
	journal.AuditRequest(fName, r, audit, journal.AuditCreate)

	if !persist.IsKEKManagerInitialized() {
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

	snapshot, err := manager.PrepareRMKRotation()
	if err != nil {
		log.Log().Error(fName, "message", "failed to create RMK snapshot", "err", err.Error())
		responseBody := net.MarshalBody(map[string]any{
			"err":     data.ErrBadInput,
			"message": "Failed to create snapshot: " + err.Error(),
		}, w)
		if responseBody == nil {
			return errors.ErrMarshalFailure
		}
		net.Respond(http.StatusInternalServerError, responseBody, w)
		return nil
	}

	response := map[string]any{
		"rmk_version": snapshot.RMKVersion,
		"kek_count":   len(snapshot.KEKs),
		"timestamp":   snapshot.Timestamp,
		"message":     "Snapshot created successfully",
	}

	responseBody := net.MarshalBody(response, w)
	if responseBody == nil {
		return errors.ErrMarshalFailure
	}

	net.Respond(http.StatusOK, responseBody, w)
	log.Log().Info(fName, "message", "created RMK rotation snapshot",
		"rmk_version", snapshot.RMKVersion,
		"kek_count", len(snapshot.KEKs))
	return nil
}
