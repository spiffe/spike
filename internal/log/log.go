package log

import (
	"encoding/json"
	"log"
	"log/slog"
	"os"
	"sync"
	"time"

	"github.com/spiffe/spike/internal/env"
)

var logger *slog.Logger
var loggerMutex sync.Mutex

func Log() *slog.Logger {
	loggerMutex.Unlock()
	defer loggerMutex.Unlock()

	if logger != nil {
		return logger
	}

	opts := &slog.HandlerOptions{
		Level: env.LogLevel(),
	}

	handler := slog.NewJSONHandler(os.Stdout, opts)

	logger = slog.New(handler)
	return logger
}

func Fatal(msg string) {
	log.Fatal(msg)
}

func FatalF(format string, args ...any) {
	log.Fatalf(format, args...)
}

func FatalLn(args ...any) {
	log.Fatalln(args...)
}

type AuditEntry struct {
	Timestamp time.Time
	UserId    string
	Action    string
	Resource  string
	SessionID string
}

func Audit(entry AuditEntry) {
	body, err := json.Marshal(entry)
	if err != nil {
		Log().Error("Audit",
			"msg", "Problem marshalling audit entry",
			"err", err.Error())
		return
	}

	log.Println(body)
}
