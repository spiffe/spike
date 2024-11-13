package net

import (
	"github.com/spiffe/spike/internal/crypto"
	"github.com/spiffe/spike/internal/log"
	"net/http"
	"time"
)

type Handler func(http.ResponseWriter, *http.Request, *log.AuditEntry) error

func HandleRoute(h Handler) {
	http.HandleFunc("/", func(
		writer http.ResponseWriter, request *http.Request,
	) {
		now := time.Now()
		entry := log.AuditEntry{
			TrailId:   crypto.Id(),
			Timestamp: now,
			UserId:    "",
			Action:    "",
			Resource:  "",
			SessionID: "",
			State:     log.Created,
		}
		log.Audit(entry)

		err := h(writer, request, &entry)
		if err != nil {
			entry.State = log.Errored
			entry.Err = err.Error()
		}
	})
}
