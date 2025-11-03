//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package kek

import (
	"net/http"
	"time"

	"github.com/spiffe/spike-sdk-go/api/entity/data"
	"github.com/spiffe/spike-sdk-go/api/errors"
	"github.com/spiffe/spike-sdk-go/log"

	"github.com/spiffe/spike/app/nexus/internal/state/persist"
	"github.com/spiffe/spike/internal/config"
	"github.com/spiffe/spike/internal/journal"
	"github.com/spiffe/spike/internal/net"
)

// RouteGetKEKStats returns statistics about KEK rotation
//
// Request: GET /v1/kek/stats
// Response: { "current_kek": {...}, "rotation_policy": {...}, "next_rotation": "..." }
func RouteGetKEKStats(
	w http.ResponseWriter, r *http.Request, audit *journal.AuditEntry,
) error {
	const fName = "routeGetKEKStats"
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

	// Calculate when rotation is needed
	rotationDays := config.KEKRotationDays()
	maxWraps := config.KEKMaxWraps()

	daysSinceCreation := time.Since(metadata.CreatedAt).Hours() / 24
	daysUntilRotation := float64(rotationDays) - daysSinceCreation
	wrapsRemaining := maxWraps - metadata.WrapsCount

	// Determine next rotation trigger
	var nextRotationReason string
	var nextRotationTime *time.Time

	if daysUntilRotation < 0 || wrapsRemaining < 0 {
		nextRotationReason = "rotation overdue"
	} else if daysUntilRotation < (float64(wrapsRemaining) / float64(maxWraps) * float64(rotationDays)) {
		nextRotationReason = "time-based"
		estimatedTime := metadata.CreatedAt.Add(time.Duration(rotationDays) * 24 * time.Hour)
		nextRotationTime = &estimatedTime
	} else {
		nextRotationReason = "usage-based"
	}

	response := map[string]any{
		"current_kek": map[string]any{
			"kek_id":      metadata.ID,
			"version":     metadata.Version,
			"created_at":  metadata.CreatedAt,
			"age_days":    int(daysSinceCreation),
			"wraps_count": metadata.WrapsCount,
			"status":      string(metadata.Status),
		},
		"rotation_policy": map[string]any{
			"rotation_days":       rotationDays,
			"max_wraps":           maxWraps,
			"grace_days":          config.KEKGraceDays(),
			"lazy_rewrap_enabled": config.KEKLazyRewrapEnabled(),
			"max_rewrap_qps":      config.KEKMaxRewrapQPS(),
		},
		"rotation_status": map[string]any{
			"days_until_rotation":  int(daysUntilRotation),
			"wraps_remaining":      wrapsRemaining,
			"next_rotation_reason": nextRotationReason,
			"next_rotation_time":   nextRotationTime,
			"rotation_needed":      manager.ShouldRotate(),
		},
	}

	responseBody := net.MarshalBody(response, w)
	if responseBody == nil {
		return errors.ErrMarshalFailure
	}

	net.Respond(http.StatusOK, responseBody, w)
	log.Log().Info(fName, "message", "returned KEK stats")
	return nil
}
