package health

import (
	"context"
	"crypto/fips140"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
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

// RouteGetStatus handles GET requests to the /v1/operator/status endpoint.
// It performs the following steps:
// Audits the incoming request for monitoring and compliance purposes.
// Guards the request by validating the SPIFFE ID and checking ACL permissions.
// Fetches the current system status, including keeper status, root key status, backing store health, FIPS mode, secrets count, and overall health.
// Marshals the aggregated system status into JSON and responds to the client.
// Handles errors gracefully, returning appropriate HTTP status codes and messages.
func RouteGetStatus(
	w http.ResponseWriter, r *http.Request, audit *journal.AuditEntry,
) error {
	const fName = "routeGetStatus"

	// Audit the request
	journal.AuditRequest(fName, r, audit, journal.AuditRead)

	err := guardStatusRequest(w, r)
	if err != nil {
		return err
	}

	status, err := getSystemStatus(r.Context())
	if err != nil {
		responseBody := net.MarshalBody(map[string]string{
			"error": "failed to get system status",
		}, w)
		net.Respond(http.StatusInternalServerError, responseBody, w)
		return err
	}

	// Marshal & respond
	responseBody := net.MarshalBody(status, w)
	if responseBody == nil {
		return apiErr.ErrMarshalFailure
	}

	net.Respond(http.StatusOK, responseBody, w)
	log.Log().Info(fName, "message", "Status returned successfully")
	return nil
}

// getSystemStatus concurrently collects status information from all critical
// system components using goroutines and a wait group. It ensures that:
//   - Keeper cluster status is retrieved.
//   - Root key availability and source are checked.
//   - Backing store connectivity and performance metrics are measured.
//   - FIPS mode and overall system health are determined.
//
// The function uses a shared mutex to safely write results from multiple
// goroutines and a context with timeout to prevent blocking forever.
// Returns a fully aggregated StatusResponse or an error if the operation
// times out.
func getSystemStatus(ctx context.Context) (StatusResponse, error) {
	// Set a reasonable timeout for the entire operation
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	type result struct {
		keepers      KeeperStatus
		rootKey      RootKeyStatus
		backingStore BackingStore
		fipsMode     bool
		secretsCount *int
		health       string
		err          error
	}

	res := result{}
	var wg sync.WaitGroup
	var mu sync.Mutex

	wg.Add(4)

	// Keeper status
	go func() {
		defer wg.Done()
		k := getKeeperStatus()
		mu.Lock()
		res.keepers = k
		mu.Unlock()
	}()

	// Root key status
	go func() {
		defer wg.Done()
		r := getRootKeyStatus()
		mu.Lock()
		res.rootKey = r
		mu.Unlock()
	}()

	// Backing store status
	go func() {
		defer wg.Done()
		b := getBackingStoreStatus()
		mu.Lock()
		res.backingStore = b
		res.secretsCount = getSecretsCount()
		mu.Unlock()
	}()

	// FIPS mode & overall health
	go func() {
		defer wg.Done()
		f := fipsMode()
		h := determineOverallHealth()
		mu.Lock()
		res.fipsMode = f
		res.health = h
		mu.Unlock()
	}()

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-ctx.Done():
		return StatusResponse{}, ctx.Err() // timeout
	case <-done:
		// Return aggregated status
		return StatusResponse{
			Health:        res.health,
			Timestamp:     time.Now(),
			Keepers:       res.keepers,
			RootKey:       res.rootKey,
			BackingStore:  res.backingStore,
			FIPSMode:      res.fipsMode,
			SecretsCount:  res.secretsCount,
			UptimeSeconds: int64(time.Since(startTime).Seconds()),
		}, nil
	}
}

// determineOverallHealth evaluates all critical system components
// to determine the overall system health status. The health determination
// follows a hierarchical approach where certain failures are more critical
// than others.
//
// Health status levels:
//   - "OK": All critical components are functioning normally
//   - "BACKING_STORE_FAILURE": Backing store is unavailable or not healthy (only for Sqlite)
//   - "ROOT_KEY_UNAVAILABLE": Root key is missing, system cannot perform crypto operations
//
// Notes:
//   - Memory mode: root key unavailable by design, system is still considered healthy
//   - Lite mode: backing store does not exist, but root key is required
//   - Sqlite mode: both backing store and root key are checked
//
// The function uses separate checks for backing store and root key to ensure
// each critical component is evaluated independently.
func determineOverallHealth() string {
	// BACKING STORE CHECK
	switch env.BackendStoreType() {
	case env.Memory, env.Lite:
		// No backing store, skip check
	default: // Sqlite
		if !backingStoreHealthy() {
			return "BACKING_STORE_FAILURE"
		}
	}

	// ROOT KEY CHECK
	switch env.BackendStoreType() {
	case env.Memory:
		// Root key unavailable by design; still healthy
	default: // Lite and Sqlite
		if !rootKeyAvailable() {
			return "ROOT_KEY_UNAVAILABLE"
		}
	}

	return "OK"
}

func backingStoreHealthy() bool {
	defer func() {
		if recover() != nil {
			// Panic occurred during storage operation, treat as unhealthy
		}
	}()

	keys := state.ListKeys()
	return keys != nil
}

func rootKeyAvailable() bool {
	switch env.BackendStoreType() {
	case env.Memory:
		// In-memory mode: no root key by design
		return false
	default:
		// Lite or Sqlite
		return !state.RootKeyZero()
	}
}

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

func getRootKeyStatus() RootKeyStatus {
	if rootKeyAvailable() {
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

func getBackingStoreStatus() BackingStore {
	start := time.Now()

	healthy := backingStoreHealthy()
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

func getBackingStoreType() string {
	backendType := env.BackendStoreType()
	log.Log().Debug("message", "Detected backing store type", "type", backendType)
	return string(backendType)
}

func getSecretsCount() *int {
	secrets := state.ListKeys()
	if backingStoreHealthy() {
		count := len(secrets)
		return &count
	}
	return nil
}

func fipsMode() bool {
	return fips140.Enabled()
}
