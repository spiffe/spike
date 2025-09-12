package health

import (
	"crypto/fips140"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	apiErr "github.com/spiffe/spike-sdk-go/api/errors"

	"github.com/spiffe/spike-sdk-go/log"
	env "github.com/spiffe/spike/app/nexus/internal/env"
	state "github.com/spiffe/spike/app/nexus/internal/state/base"
	"github.com/spiffe/spike/internal/journal"
	"github.com/spiffe/spike/internal/net"
)

// TODO(doguhannilt): Move StatusResponse and related structs to spike-sdk-go/api/data/status.go
// These are part of the public API contract and should live in the SDK.

// StatusResponse represents the complete system status information
// returned by the /v1/operator/status endpoint. It provides a comprehensive
// view of all critical system components including health status, keeper
// availability, root key status, and backing store connectivity.
type StatusResponse struct {
	Health        string        `json:"health"`
	Timestamp     time.Time     `json:"timestamp"`
	Keepers       KeeperStatus  `json:"keepers"`
	RootKey       RootKeyStatus `json:"root_key"`
	BackingStore  BackingStore  `json:"backing_store"`
	FIPSMode      bool          `json:"fips_mode"`
	SecretsCount  *int          `json:"secrets_count,omitempty"`
	UptimeSeconds int64         `json:"uptime_seconds"`
}

// KeeperStatus represents the status of the keeper cluster used for
// Shamir secret sharing. It tracks both the configured keeper instances
// and the threshold required for root key reconstruction.
type KeeperStatus struct {
	Status        string `json:"status"`
	ActiveCount   int    `json:"active_count"`
	RequiredCount int    `json:"required_count"`
}

// RootKeyStatus indicates whether the root encryption key is available
// for cryptographic operations and where it's sourced from.
type RootKeyStatus struct {
	Status string `json:"status"`
	Source string `json:"source"`
}

// BackingStore represents the connection status and performance metrics
// of the persistent storage backend used for secret storage.
type BackingStore struct {
	Status         string `json:"status"`
	Type           string `json:"type"`
	ResponseTimeMs *int   `json:"response_time_ms,omitempty"`
}

// startTime records when the server was started, used for uptime calculation
var startTime = time.Now()

// RouteGetStatus handles GET requests to /v1/operator/status endpoint.
// It returns a comprehensive JSON response containing the current status
// of all critical system components including health, keepers, root key,
// backing store, FIPS mode, secrets count, and uptime.
//
// The endpoint provides essential monitoring information for:
//   - Health check systems and load balancers
//   - Prometheus/Grafana monitoring and alerting
//   - Kubernetes readiness and liveness probes
//   - Operations teams for system diagnostics
//
// Response format conforms to standard REST API patterns with proper
// HTTP status codes and JSON content type headers.
func RouteGetStatus(
	w http.ResponseWriter, r *http.Request, audit *journal.AuditEntry,
) error {
	const fName = "routeGetStatus"

	// 1️⃣ Audit the request
	journal.AuditRequest(fName, r, audit, journal.AuditRead)

	err := guardStatusRequest(w, r)
	if err != nil {
		return err
	}

	// 3️⃣ Get system status
	status := getSystemStatus()

	// 4️⃣ Marshal & respond
	responseBody := net.MarshalBody(status, w)
	if responseBody == nil {
		return apiErr.ErrMarshalFailure
	}

	net.Respond(http.StatusOK, responseBody, w)
	log.Log().Info(fName, "message", "Status returned successfully")
	return nil
}

// getSystemStatus collects and aggregates status information from all
// critical system components. This function coordinates health checks
// across keepers, root key availability, backing store connectivity,
// and system configuration to provide a complete status snapshot.
//
// Returns a StatusResponse containing all relevant system metrics
// and health indicators at the current point in time.
func getSystemStatus() StatusResponse {
	return StatusResponse{
		Health:        determineOverallHealth(),
		Timestamp:     time.Now(),
		Keepers:       getKeeperStatus(),
		RootKey:       getRootKeyStatus(),
		BackingStore:  getBackingStoreStatus(),
		FIPSMode:      isFIPSMode(),
		SecretsCount:  getSecretsCount(),
		UptimeSeconds: int64(time.Since(startTime).Seconds()),
	}
}

// determineOverallHealth evaluates all critical system components
// to determine the overall system health status. The health determination
// follows a hierarchical approach where certain failures are more critical
// than others.
//
// Health status levels:
//   - "OK": All critical components are functioning normally
//   - "DEGRADED": Backing store issues but system partially functional
//   - "UNAVAILABLE": Root key unavailable, system cannot perform crypto operations
//
// The function prioritizes root key availability over backing store health
// since cryptographic operations are impossible without the root key.
func determineOverallHealth() string {
	// Check all critical components
	if !isBackingStoreHealthy() {
		return "DEGRADED"
	}

	if !isRootKeyAvailable() {
		return "UNAVAILABLE"
	}

	return "OK"
}

// getKeeperStatus examines the keeper cluster configuration and determines
// the health of the Shamir secret sharing infrastructure. It reads keeper
// peer URLs from environment variables and compares the available count
// against the required threshold.
//
// The function parses SPIKE_NEXUS_KEEPER_PEERS environment variable to
// determine configured keeper instances and SPIKE_NEXUS_SHAMIR_THRESHOLD
// to establish minimum requirements for root key reconstruction.
//
// Status determination:
//   - "HEALTHY": Sufficient keepers available to meet threshold requirements
//   - "DEGRADED": Some keepers configured but below threshold
//   - "UNHEALTHY": No keepers configured
//   - "UNKNOWN": Configuration missing or invalid
//
// Note: This performs configuration-level checks only, not actual
// network connectivity tests to keeper instances.
func getKeeperStatus() KeeperStatus {

	peersEnv := os.Getenv("SPIKE_NEXUS_KEEPER_PEERS")
	if peersEnv == "" {
		return KeeperStatus{
			Status:        "UNKNOWN",
			ActiveCount:   0,
			RequiredCount: 0,
		}
	}

	urls := strings.Split(peersEnv, ",")
	cleanURLs := make([]string, 0, len(urls))
	for _, u := range urls {
		trimmed := strings.TrimSpace(u)
		if trimmed != "" {
			cleanURLs = append(cleanURLs, trimmed)
		}
	}

	activeCount := len(cleanURLs)
	requiredCountStr := os.Getenv("SPIKE_NEXUS_SHAMIR_THRESHOLD")
	requiredCount, err := strconv.Atoi(requiredCountStr)
	if err != nil || requiredCount <= 0 {
		requiredCount = 2
	}

	status := "HEALTHY"
	if activeCount == 0 {
		status = "UNHEALTHY"
	} else if activeCount < requiredCount {
		status = "DEGRADED"
	}

	return KeeperStatus{
		Status:        status,
		ActiveCount:   activeCount,
		RequiredCount: requiredCount,
	}
}

// getRootKeyStatus determines the availability and source of the root
// encryption key used for all cryptographic operations. The root key
// is essential for secret encryption, decryption, and system functionality.
//
// The function checks both the key's presence in memory and the availability
// of sufficient keeper instances to reconstruct the key if needed.
//
// Status values:
//   - "AVAILABLE": Root key is present and usable for crypto operations
//   - "UNAVAILABLE": Root key missing, zero, or insufficient keepers
//
// Source values:
//   - "keeper": Root key available through keeper reconstruction
//   - "unknown": Root key source cannot be determined
func getRootKeyStatus() RootKeyStatus {
	if isRootKeyAvailable() {
		return RootKeyStatus{
			Status: "AVAILABLE",
			Source: "keeper",
		}
	}

	return RootKeyStatus{
		Status: "UNAVAILABLE",
		Source: "unknown",
	}
}

// getBackingStoreStatus tests the connectivity and performance of the
// persistent storage backend. It measures response time and determines
// connection health through basic operations.
//
// The function performs a lightweight health check and measures the
// time taken to complete the operation, providing both connectivity
// status and performance metrics.
//
// Status values:
//   - "CONNECTED": Backing store is reachable and responsive
//   - "DISCONNECTED": Backing store unavailable or unresponsive
//
// Response time is measured in milliseconds and included in the status
// for performance monitoring and alerting purposes.
func getBackingStoreStatus() BackingStore {
	start := time.Now()

	healthy := isBackingStoreHealthy()
	responseTime := int(time.Since(start).Milliseconds())

	status := "CONNECTED"
	if !healthy {
		status = "DISCONNECTED"
	}

	return BackingStore{
		Status:         status,
		Type:           getBackingStoreType(),
		ResponseTimeMs: &responseTime,
	}
}

// getBackingStoreType identifies the type of persistent storage backend
// currently configured for the system. This information helps operations
// teams understand the storage architecture and troubleshoot issues.
//
// The function queries the environment configuration to determine the
// backend type and logs the detected type for debugging purposes.
//
// Common backend types include:
//   - "postgres": PostgreSQL database backend
//   - "mysql": MySQL database backend
//   - "file": File system backend
//   - "memory": In-memory backend (testing/development)
func getBackingStoreType() string {
	backendType := env.BackendStoreType()

	log.Log().Debug("message", "Detected backing store type", "type", backendType)
	return string(backendType)
}

// getSecretsCount returns the total number of secrets currently stored
// in the backing store. This metric is useful for capacity planning,
// monitoring system usage, and detecting unexpected changes in secret count.
//
// The function only returns a count when the backing store is healthy
// to ensure the reported number is accurate and reliable. If the backing
// store is unhealthy, nil is returned to indicate the count cannot be
// determined.
//
// Returns:
//   - *int: Pointer to secret count when backing store is healthy
//   - nil: When backing store is unhealthy or inaccessible
func getSecretsCount() *int {
	secrets := state.ListKeys()
	if isBackingStoreHealthy() {
		count := len(secrets)
		return &count
	}
	return nil
}

// isFIPSMode reports whether the system is currently operating in
// FIPS 140-3 compliance mode. This is critical information for
// environments that require cryptographic compliance with federal
// standards.
//
// The function uses Go's official crypto/fips140 package to determine
// the current FIPS status, which is set via GODEBUG environment
// variables at program startup.
//
// Returns:
//   - true: System is operating in FIPS 140-3 compliant mode
//   - false: System is using standard cryptographic implementations
func isFIPSMode() bool {
	return fips140.Enabled()
}

// isRootKeyAvailable performs a comprehensive check to determine if
// the root encryption key is available for cryptographic operations.
// This is one of the most critical system health indicators.
//
// The function performs two essential checks:
//  1. Verifies the root key is present in memory (not zero)
//  2. Ensures sufficient keeper instances are available for key reconstruction
//
// Both conditions must be satisfied for the root key to be considered
// available, as a zero key indicates an uninitialized system, and
// insufficient keepers prevent key reconstruction during recovery scenarios.
//
// Returns:
//   - true: Root key is available and system can perform crypto operations
//   - false: Root key unavailable, system cannot encrypt/decrypt secrets
func isRootKeyAvailable() bool {
	if state.RootKeyZero() {
		return false
	}

	keeperStatus := getKeeperStatus()
	if keeperStatus.ActiveCount < keeperStatus.RequiredCount {
		return false
	}

	return true
}

// isBackingStoreHealthy performs a basic connectivity test to the
// persistent storage backend. This function provides a lightweight
// health check that can be called frequently without significant
// performance impact.
//
// The health check attempts to perform a simple read operation
// (listing keys) and handles any panics that might occur due to
// connection failures, timeouts, or other storage-related issues.
//
// The function uses a defensive programming approach with panic
// recovery to ensure that storage backend failures don't crash
// the status endpoint, allowing the system to report degraded
// status rather than becoming completely unavailable.
//
// Returns:
//   - true: Backing store is responsive and accessible
//   - false: Backing store is unreachable, timed out, or failed
func isBackingStoreHealthy() bool {
	// Defensive health check with panic recovery
	defer func() {
		if recover() != nil {
			// Panic occurred during storage operation
			// This indicates an unhealthy backing store
		}
	}()

	keys := state.ListKeys()
	return keys != nil
}
