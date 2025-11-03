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

// RouteGetCurrentKEK returns information about the current active KEK
//
// Request: GET /v1/kek/current
// Response: { "kek_id": "v1-2025-01", "version": 1, "created_at": "...", ... }
func RouteGetCurrentKEK(
	w http.ResponseWriter, r *http.Request, audit *journal.AuditEntry,
) error {
	const fName = "routeGetCurrentKEK"
	journal.AuditRequest(fName, r, audit, journal.AuditRead)

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

	currentKekID := manager.GetCurrentKEKID()
	metadata, err := manager.GetMetadata(currentKekID)
	if err != nil {
		log.Log().Error(fName, "message", "failed to get KEK metadata", "err", err.Error())
		responseBody := net.MarshalBody(map[string]any{
			"err":     data.ErrBadInput,
			"message": "Failed to get KEK metadata",
		}, w)
		if responseBody == nil {
			return errors.ErrMarshalFailure
		}
		net.Respond(http.StatusInternalServerError, responseBody, w)
		return nil
	}

	response := map[string]any{
		"kek_id":      metadata.ID,
		"version":     metadata.Version,
		"created_at":  metadata.CreatedAt,
		"wraps_count": metadata.WrapsCount,
		"status":      string(metadata.Status),
		"rmk_version": metadata.RMKVersion,
	}

	responseBody := net.MarshalBody(response, w)
	if responseBody == nil {
		return errors.ErrMarshalFailure
	}

	net.Respond(http.StatusOK, responseBody, w)
	log.Log().Info(fName, "message", "returned current KEK info", "kek_id", currentKekID)
	return nil
}
