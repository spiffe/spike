package health

import (
	"context"
	"crypto/fips140"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/spiffe/go-spiffe/v2/workloadapi"
	"github.com/spiffe/spike-sdk-go/api/entity/data"
	apiErr "github.com/spiffe/spike-sdk-go/api/errors"

	"github.com/spiffe/spike-sdk-go/api/entity/v1/reqres"
	"github.com/spiffe/spike-sdk-go/log"
	env "github.com/spiffe/spike/app/nexus/internal/env"
	state "github.com/spiffe/spike/app/nexus/internal/state/base"
	"github.com/spiffe/spike/internal/journal"
	"github.com/spiffe/spike/internal/net"
)

// startTime records when the server was started, used for uptime calculation
var startTime = time.Now()

// RouteGetStatus handles GET requests to the /v1/operator/status endpoint.
//
// It performs the following steps:
//   - Audits the incoming request for monitoring and compliance purposes.
//   - Guards the request by validating the SPIFFE ID and checking ACL
//     permissions.
//   - Fetches the current system status, including keeper status, root key
//     status, backing store health, FIPS mode, secrets count, and overall
//     health.
//   - Marshals the aggregated system status into JSON and responds to the
//     client.
//   - Handles errors gracefully, returning appropriate HTTP status codes and
//     messages.
func RouteGetStatus(
	w http.ResponseWriter, r *http.Request, audit *journal.AuditEntry,
) error {
	const fName = "routeGetStatus"
	// Audit
	journal.AuditRequest(fName, r, audit, journal.AuditRead)

	requestBody := net.ReadRequestBody(w, r)

	// HandleRequest pattern
	request := net.HandleRequest[
		reqres.HealthReadRequest, reqres.HealthReadResponse](
		requestBody, w,
		reqres.HealthReadResponse{Err: data.ErrBadInput}, // default error
	)

	if request == nil {
		request = &reqres.HealthReadRequest{
			Version: 0,
		}
	}

	err := guardStatusRequest(w, r)
	if err != nil {
		return err
	}

	status, err := getSystemStatus(r.Context())
	if err != nil {
		responseBody := net.MarshalBody(reqres.HealthReadResponse{
			StatusResponse: data.StatusResponse{},
			Err:            data.ErrBadInput,
		}, w)
		net.Respond(http.StatusInternalServerError, responseBody, w)
		return err
	}

	// Marshal & respond
	responseBody := net.MarshalBody(reqres.HealthReadResponse{
		StatusResponse: status,
		Err:            "",
	}, w)
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
func getSystemStatus(ctx context.Context) (data.StatusResponse, error) {

	// Set a reasonable timeout for the entire operation
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Result struct to hold intermediate results
	res := data.Result{}

	var wg sync.WaitGroup
	var mu sync.Mutex

	wg.Add(3) // 3 concurrent tasks

	// Root key status
	go func() {
		defer wg.Done()
		r := getRootKeyStatus()
		mu.Lock()
		res.RootKey = r
		mu.Unlock()
	}()

	// Backing store status
	go func() {
		defer wg.Done()
		b := getBackingStoreStatus()
		mu.Lock()
		res.BackingStore = b
		res.SecretsCount = getSecretsCount()
		mu.Unlock()
	}()

	// FIPS mode & overall health
	go func() {
		defer wg.Done()
		f := fipsMode()
		h := determineOverallHealth()
		mu.Lock()
		res.FipsMode = f
		res.Health = h
		mu.Unlock()
	}()

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-ctx.Done():
		return data.StatusResponse{}, ctx.Err() // timeout
	case <-done:
		// Return aggregated status
		return data.StatusResponse{
			Health:        res.Health,
			Timestamp:     time.Now(),
			RootKey:       res.RootKey,
			BackingStore:  res.BackingStore,
			FIPSMode:      res.FipsMode,
			SecretsCount:  res.SecretsCount,
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

// backingStoreHealthy performs a simple write-read-delete operation
// on the backing store to verify its health and responsiveness.
// It returns true if all operations succeed, indicating the backing
func backingStoreHealthy() bool {
	defer func() {
		if recover() != nil {
			// Panic occurred during storage operation, treat as unhealthy
		}
	}()

	key := "healthcheck-temp-key"
	values := map[string]string{"value": "test"}

	// Set a temporary secret
	err := state.UpsertSecret(key, values)
	if err != nil {
		fmt.Println("Failed to upsert secret:", err)
		return false
	}

	// Get the secret back
	vals, err := state.GetSecret(key, 0)
	if err != nil {
		fmt.Println("Failed to get secret:", err)
		_ = state.DeleteSecret(key, nil)
		return false
	}
	if v, ok := vals["value"]; !ok || v != "test" {
		fmt.Println("Secret value mismatch")
		_ = state.DeleteSecret(key, nil)
		return false
	}

	// Delete the secret
	err = state.DeleteSecret(key, nil) // nil = delete current version only
	if err != nil {
		fmt.Println("Failed to delete secret:", err)
		return false
	}

	return true
}

// rootKeyAvailable checks if the root key is available in the system.
// The availability depends on the backend store type:
//   - Memory: root key is unavailable by design
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

// getRootKeyStatus checks if the root key is available and returns its status.
func getRootKeyStatus() data.RootKeyStatus {
	if rootKeyAvailable() {
		return data.RootKeyStatus{
			Status: "AVAILABLE",
		}
	}

	return data.RootKeyStatus{
		Status: "UNAVAILABLE",
	}
}

// getBackingStoreStatus checks the health and type of the backing store.
// It performs a simple write-read-delete operation to verify connectivity
// and responsiveness. The function returns the backing store status along
// with its type and response time in milliseconds.
func getBackingStoreStatus() data.BackingStore {
	start := time.Now()

	healthy := backingStoreHealthy()
	responseTime := int(time.Since(start).Milliseconds())

	status := "CONNECTED"
	if !healthy {
		status = "DISCONNECTED"
	}

	return data.BackingStore{
		Status:         status,
		Type:           getBackingStoreType(),
		ResponseTimeMs: &responseTime,
	}
}

// getbackingStoreType retrieves the type of the backing store from
// environment configuration. It logs the detected type for debugging
// purposes and returns it as a string.
func getBackingStoreType() string {
	backendType := env.BackendStoreType()
	return string(backendType)
}

// getSecretsCount retrieves the total number of secrets stored in the backing store.
// It returns nil if the store type is Memory or Lite, as these modes do not
// persist secrets. If the backing store is not healthy, it also returns nil.
func getSecretsCount() *int {
	storeType := env.BackendStoreType()

	if storeType == env.Lite || storeType == env.Memory {
		return nil
	}

	if !backingStoreHealthy() {
		return nil
	}

	secrets := state.ListKeys()
	count := len(secrets)
	return &count
}

// fipsMode checks if the system is running in FIPS 140 mode.
func fipsMode() bool {
	return fips140.Enabled()
}

// getSpireAgentSocketPath retrieves the SPIRE agent socket path from the
// SPIRE_AGENT_SOCKET environment variable. If not set, it defaults to
// /tmp/spire-agent/public/api.sock. The function returns the path in the
// Unix socket URI format required by the workload API client.
// Experimental: not used currently.
func getSpireAgentSocketPath() string {
	path := os.Getenv("SPIRE_AGENT_SOCKET")
	if path == "" {
		path = "/tmp/spire-agent/public/api.sock"
	}
	// Unix socket URI formatı: unix:///tmp/spire-agent/public/api.sock
	return "unix://" + path
}

// -----------------------------------------------------------------------------
// Keeper Status (Experimental / Not Used)
// -----------------------------------------------------------------------------
//
// The following code implements health checks for SPIKE Keeper nodes in the cluster.
// It connects to each keeper's health endpoint using mTLS and aggregates their status.
//
// NOTE: This functionality is currently **not used** in the status response and is
// under development. You can safely ignore or move this code for now.
// It will be integrated into the main status endpoint in a future release.
//
// -----------------------------------------------------------------------------
// getKeeperStatus checks the health of all configured keepers in the cluster.
// It attempts to connect to each keeper's health endpoint using the provided
// X.509 source for mTLS authentication. The function counts how many keepers
func getKeeperStatus(source *workloadapi.X509Source) data.KeeperStatus {
	keepers := env.Keepers()
	activeCount := 0

	if len(keepers) == 0 {
		return data.KeeperStatus{
			Status:      "NO_KEEPERS_CONFIGURED",
			ActiveCount: activeCount,
		}
	}

	for _, keeperURL := range keepers {
		svid, err := source.GetX509SVID()
		if err != nil {
			continue
		}

		cert := tls.Certificate{
			Certificate: make([][]byte, len(svid.Certificates)),
			PrivateKey:  svid.PrivateKey,
		}
		for i, c := range svid.Certificates {
			cert.Certificate[i] = c.Raw
		}

		resp, err := callKeeperHealth(keeperURL, cert)
		if err != nil {
			fmt.Printf("[ERROR] Keeper not reachable: %s, err: %v\n", keeperURL, err)
			continue
		}
		if resp.StatusCode == http.StatusOK {
			activeCount++
		}
		resp.Body.Close()
	}

	status := "UNHEALTHY"
	if activeCount > 0 {
		status = "HEALTHY"
	}

	fmt.Printf("[DEBUG] Keeper status result: %s (active=%d)\n", status, activeCount)

	return data.KeeperStatus{
		Status:      status,
		ActiveCount: activeCount,
	}
}

// callKeeperHealth performs a health check request to a keeper node.
// NOTE: This function is part of the experimental keeper status logic.
func callKeeperHealth(keeperURL string, cert tls.Certificate) (*http.Response, error) {
	fmt.Printf("[DEBUG] Sending health check to: %s/v1/store/health\n", keeperURL)

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				Certificates:       []tls.Certificate{cert},
				InsecureSkipVerify: true, // For demo purposes only; use proper cert verification in production
			},
		},
		Timeout: 3 * time.Second,
	}

	resp, err := client.Get(keeperURL + "/v1/store/health?action=")
	if err != nil {
		fmt.Printf("[ERROR] HTTP request failed: %v\n", err)
	} else {
		fmt.Printf("[DEBUG] Keeper %s responded with status: %d\n", keeperURL, resp.StatusCode)
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("[DEBUG] Keeper response body: %s\n", string(body))
	}
	return resp, err
}
