package net

import (
	"net/http"
	"time"

	"github.com/spiffe/spike/internal/log"
	"github.com/spiffe/spike/pkg/crypto"
)

// TODO: document
type Handler func(http.ResponseWriter, *http.Request, *log.AuditEntry) error

// TODO: document
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
