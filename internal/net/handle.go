package net

import (
	"net/http"
	"time"

	"github.com/spiffe/spike/internal/log"
	"github.com/spiffe/spike/pkg/crypto"
)

// Handler is a function type that processes HTTP requests with audit
// logging support.
type Handler func(http.ResponseWriter, *http.Request, *log.AuditEntry) error

// HandleRoute wraps an HTTP handler with audit logging functionality.
// It creates and manages audit log entries for the request lifecycle,
// including:
// - Generating unique trail IDs
// - Recording timestamps and durations
// - Tracking request status (created, success, error)
// - Capturing error information
//
// The wrapped handler is mounted at the root path ("/") and automatically
// logs entry and exit audit events for all requests.
//
// Parameters:
//   - h: Handler function to wrap with audit logging
func HandleRoute(h Handler) {
	http.HandleFunc("/", func(
		writer http.ResponseWriter, request *http.Request,
	) {
		now := time.Now()
		entry := log.AuditEntry{
			TrailId:   crypto.Id(),
			Timestamp: now,
			UserId:    "",
			Action:    log.AuditEnter,
			Path:      request.URL.Path,
			Resource:  "",
			SessionID: "",
			State:     log.AuditCreated,
		}
		log.Audit(entry)

		err := h(writer, request, &entry)
		if err == nil {
			entry.State = log.AuditSuccess
		} else {
			entry.State = log.AuditErrored
			entry.Err = err.Error()
		}

		entry.Duration = time.Since(now)
		log.Audit(entry)
	})
}
