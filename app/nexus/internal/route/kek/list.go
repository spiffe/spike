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

// RouteListKEKs returns a list of all KEKs with their metadata
//
// Request: GET /v1/kek/list
// Response: { "keks": [ { "kek_id": "...", "version": 1, ... }, ... ] }
func RouteListKEKs(
	w http.ResponseWriter, r *http.Request, audit *journal.AuditEntry,
) error {
	const fName = "routeListKEKs"
	journal.AuditRequest(fName, r, audit, journal.AuditList)

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

	// Get all KEK metadata from storage
	allMetadata, err := manager.ListAllKEKs()
	if err != nil {
		log.Log().Error(fName, "message", "failed to list KEKs", "err", err.Error())
		responseBody := net.MarshalBody(map[string]any{
			"err":     data.ErrBadInput,
			"message": "Failed to list KEKs: " + err.Error(),
		}, w)
		if responseBody == nil {
			return errors.ErrMarshalFailure
		}
		net.Respond(http.StatusInternalServerError, responseBody, w)
		return nil
	}

	// Convert metadata to response format
	keks := make([]map[string]any, 0, len(allMetadata))
	for _, metadata := range allMetadata {
		keks = append(keks, map[string]any{
			"kek_id":      metadata.ID,
			"version":     metadata.Version,
			"created_at":  metadata.CreatedAt,
			"wraps_count": metadata.WrapsCount,
			"status":      string(metadata.Status),
			"rmk_version": metadata.RMKVersion,
			"retired_at":  metadata.RetiredAt,
		})
	}

	response := map[string]any{
		"keks":  keks,
		"count": len(keks),
	}

	responseBody := net.MarshalBody(response, w)
	if responseBody == nil {
		return errors.ErrMarshalFailure
	}

	net.Respond(http.StatusOK, responseBody, w)
	log.Log().Info(fName, "message", "listed KEKs", "count", len(keks))
	return nil
}
