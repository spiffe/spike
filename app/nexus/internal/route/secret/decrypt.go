package secret

import (
	"fmt"
	"io"
	"net/http"

	"github.com/spiffe/spike-sdk-go/api/errors"

	"github.com/spiffe/spike/app/nexus/internal/state/persist"
	"github.com/spiffe/spike/internal/log"
	"github.com/spiffe/spike/internal/net"
)

func RouteDecrypt(
	w http.ResponseWriter, r *http.Request, audit *log.AuditEntry,
) error {
	const fName = "routeDecrypt"
	c := persist.Backend().GetCipher()
	log.AuditRequest(fName, r, audit, log.AuditCreate)

	ver := make([]byte, 1)
	n, err := io.ReadFull(r.Body, ver)
	if err != nil || n != 1 {
		log.Log().Debug(fName, "msg", "Failed to read version")
		return fmt.Errorf("Failed to read version")
	}
	if ver[0] != byte('1') {
		return fmt.Errorf("Unknown file type")
	}

	nBytesToRead := c.NonceSize()
	nonce := make([]byte, nBytesToRead)
	n, err = io.ReadFull(r.Body, nonce)
	if err != nil || n != nBytesToRead {
		log.Log().Debug(fName, "msg", "Failed to read nonce")
		return fmt.Errorf("Failed to read nonce")
	}

	requestBody := net.ReadRequestBody(w, r)
	if requestBody == nil {
		return errors.ErrReadFailure
	}

	log.Log().Info(fName, "msg", fmt.Sprintf("Decrypt %d %d", len(nonce), len(requestBody)))

	plaintext, err := c.Open(nil, nonce, requestBody, nil)
	if err != nil {
		log.Log().Info(fName, "msg", fmt.Errorf("failed to decrypt %w", err))
		return err
	}

	fmt.Fprintf(w, "%s", plaintext)
	return nil
}
