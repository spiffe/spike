package store

import (
	"encoding/json"
	"net/http"

	"github.com/spiffe/spike-sdk-go/log"
	"github.com/spiffe/spike-sdk-go/spiffe"

	"github.com/spiffe/spike/internal/net"

	"github.com/spiffe/spike/internal/journal"
)

type HealthResponse struct {
	Status string `json:"status"`
}

func RouteHealth(w http.ResponseWriter, r *http.Request, audit *journal.AuditEntry) error {
	const fName = "routeHealth"
	journal.AuditRequest(fName, r, audit, journal.AuditRead)

	ctx := r.Context()
	source, _, err := spiffe.Source(ctx, spiffe.EndpointSocket())
	if err != nil {
		log.Log().Error(fName, "message", "Cannot get SPIFFE source", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return err
	}
	defer spiffe.CloseSource(source)

	resp := HealthResponse{
		Status: "healthy",
	}

	jsonBody, _ := json.Marshal(resp)
	net.Respond(http.StatusOK, jsonBody, w)

	return nil
}
