//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package kek

import (
	"time"

	"github.com/spiffe/spike-sdk-go/log"
)

// Metrics tracks KEK-related operations for observability
type Metrics struct {
	// TotalWraps tracks total DEK wrap operations
	TotalWraps int64

	// TotalUnwraps tracks total DEK unwrap operations
	TotalUnwraps int64

	// TotalRewraps tracks total DEK rewrap operations
	TotalRewraps int64

	// TotalRotations tracks total KEK rotations
	TotalRotations int64

	// LastRotation is the timestamp of the last KEK rotation
	LastRotation time.Time

	// LastError is the last error encountered
	LastError string

	// LastErrorTime is when the last error occurred
	LastErrorTime time.Time
}

var globalMetrics = &Metrics{}

// RecordWrap records a DEK wrap operation
func RecordWrap(kekID string, success bool) {
	globalMetrics.TotalWraps++
	
	if success {
		log.Log().Info("RecordWrap",
			"kek_id", kekID,
			"total_wraps", globalMetrics.TotalWraps)
	} else {
		globalMetrics.LastError = "wrap failed"
		globalMetrics.LastErrorTime = time.Now()
		log.Log().Error("RecordWrap",
			"kek_id", kekID,
			"message", "wrap operation failed")
	}
}

// RecordUnwrap records a DEK unwrap operation
func RecordUnwrap(kekID string, success bool) {
	globalMetrics.TotalUnwraps++
	
	if success {
		log.Log().Info("RecordUnwrap",
			"kek_id", kekID,
			"total_unwraps", globalMetrics.TotalUnwraps)
	} else {
		globalMetrics.LastError = "unwrap failed"
		globalMetrics.LastErrorTime = time.Now()
		log.Log().Error("RecordUnwrap",
			"kek_id", kekID,
			"message", "unwrap operation failed")
	}
}

// RecordRewrap records a DEK rewrap operation
func RecordRewrap(oldKekID, newKekID string, success bool) {
	globalMetrics.TotalRewraps++
	
	if success {
		log.Log().Info("RecordRewrap",
			"old_kek_id", oldKekID,
			"new_kek_id", newKekID,
			"total_rewraps", globalMetrics.TotalRewraps)
	} else {
		globalMetrics.LastError = "rewrap failed"
		globalMetrics.LastErrorTime = time.Now()
		log.Log().Error("RecordRewrap",
			"old_kek_id", oldKekID,
			"new_kek_id", newKekID,
			"message", "rewrap operation failed")
	}
}

// RecordRotation records a KEK rotation
func RecordRotation(oldKekID, newKekID string, duration time.Duration, success bool) {
	if success {
		globalMetrics.TotalRotations++
		globalMetrics.LastRotation = time.Now()
		
		log.Log().Info("RecordRotation",
			"old_kek_id", oldKekID,
			"new_kek_id", newKekID,
			"duration_ms", duration.Milliseconds(),
			"total_rotations", globalMetrics.TotalRotations)
	} else {
		globalMetrics.LastError = "rotation failed"
		globalMetrics.LastErrorTime = time.Now()
		
		log.Log().Error("RecordRotation",
			"old_kek_id", oldKekID,
			"message", "rotation failed",
			"duration_ms", duration.Milliseconds())
	}
}

// GetMetrics returns a copy of current metrics
func GetMetrics() Metrics {
	return *globalMetrics
}

// ResetMetrics resets all metrics (for testing)
func ResetMetrics() {
	globalMetrics = &Metrics{}
}

// LogKEKStats logs current KEK statistics
func LogKEKStats(m *Manager) {
	const fName = "LogKEKStats"
	
	keks, err := m.ListAllKEKs()
	if err != nil {
		log.Log().Error(fName,
			"message", "failed to list KEKs",
			"err", err.Error())
		return
	}
	
	statusCounts := make(map[KekStatus]int)
	var totalWraps int64
	var oldestKEK time.Time
	var newestKEK time.Time
	
	for _, kek := range keks {
		statusCounts[kek.Status]++
		totalWraps += kek.WrapsCount
		
		if oldestKEK.IsZero() || kek.CreatedAt.Before(oldestKEK) {
			oldestKEK = kek.CreatedAt
		}
		if newestKEK.IsZero() || kek.CreatedAt.After(newestKEK) {
			newestKEK = kek.CreatedAt
		}
	}
	
	log.Log().Info(fName,
		"total_keks", len(keks),
		"active_keks", statusCounts[KekStatusActive],
		"grace_keks", statusCounts[KekStatusGrace],
		"retired_keks", statusCounts[KekStatusRetired],
		"total_wraps", totalWraps,
		"oldest_kek_age_days", int(time.Since(oldestKEK).Hours()/24),
		"metrics_total_wraps", globalMetrics.TotalWraps,
		"metrics_total_unwraps", globalMetrics.TotalUnwraps,
		"metrics_total_rewraps", globalMetrics.TotalRewraps,
		"metrics_total_rotations", globalMetrics.TotalRotations)
}

