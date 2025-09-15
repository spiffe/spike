package store

import (
	"crypto/fips140"
	"encoding/json"
	"net/http"

	"github.com/spiffe/spike-sdk-go/log"
	"github.com/spiffe/spike-sdk-go/security/mem"
	"github.com/spiffe/spike-sdk-go/spiffe"
	"github.com/spiffe/spike-sdk-go/spiffeid"

	"github.com/spiffe/spike/app/keeper/internal/env"
	"github.com/spiffe/spike/internal/net"

	"github.com/spiffe/spike/internal/journal"
)

type HealthResponse struct {
	Status        string `json:"status"`
	MemLocked     bool   `json:"mem_locked"`
	FIPS          bool   `json:"fips_enabled"`
	ValidSPIFFEID bool   `json:"valid_spiffe_id"`
}

func RouteHealth(w http.ResponseWriter, r *http.Request, audit *journal.AuditEntry) error {
	const fName = "routeHealth"
	journal.AuditRequest(fName, r, audit, journal.AuditRead)

	ctx := r.Context()
	source, selfSPIFFEID, err := spiffe.Source(ctx, spiffe.EndpointSocket())
	if err != nil {
		log.Log().Error(fName, "message", "Cannot get SPIFFE source", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return err
	}
	defer spiffe.CloseSource(source)

	validSPIFFE := spiffeid.IsKeeper(env.TrustRoot(), selfSPIFFEID)

	resp := HealthResponse{
		Status:        "healthy",
		MemLocked:     mem.Lock(),
		FIPS:          fips140.Enabled(),
		ValidSPIFFEID: validSPIFFE,
	}

	statusCode := http.StatusOK
	if !resp.MemLocked || !resp.ValidSPIFFEID {
		resp.Status = "unhealthy"
		statusCode = http.StatusServiceUnavailable
	}

	jsonBody, _ := json.Marshal(resp)
	net.Respond(statusCode, jsonBody, w)

	log.Log().Info(fName, "message", "Keeper health check executed")
	return nil
}
