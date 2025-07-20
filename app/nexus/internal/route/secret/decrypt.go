package secret

import (
	"fmt"
	"io"
	"net/http"

	"github.com/spiffe/spike-sdk-go/api/errors"
	"github.com/spiffe/spike-sdk-go/log"

	"github.com/spiffe/spike/app/nexus/internal/state/persist"
	journal "github.com/spiffe/spike/internal/log"
	"github.com/spiffe/spike/internal/net"
)

func RouteDecrypt(
	w http.ResponseWriter, r *http.Request, audit *journal.AuditEntry,
) error {
	const fName = "routeDecrypt"
	c := persist.Backend().GetCipher()
	journal.AuditRequest(fName, r, audit, journal.AuditCreate)

	ver := make([]byte, 1)
	n, err := io.ReadFull(r.Body, ver)
	if err != nil || n != 1 {
		log.Log().Debug(fName, "message", "Failed to read version")
		return fmt.Errorf("failed to read version")
	}
	if ver[0] != byte('1') {
		return fmt.Errorf("unknown file type")
	}

	bytesToRead := c.NonceSize()
	nonce := make([]byte, bytesToRead)
	n, err = io.ReadFull(r.Body, nonce)
	if err != nil || n != bytesToRead {
		log.Log().Debug(fName, "message", "Failed to read nonce")
		return fmt.Errorf("failed to read nonce")
	}

	requestBody := net.ReadRequestBody(w, r)
	if requestBody == nil {
		return errors.ErrReadFailure
	}

	log.Log().Info(fName, "message",
		fmt.Sprintf("Decrypt %d %d", len(nonce), len(requestBody)),
	)

	plaintext, err := c.Open(nil, nonce, requestBody, nil)
	if err != nil {
		log.Log().Info(fName, "message", fmt.Errorf("failed to decrypt %w", err))
		return err
	}

	_, err = fmt.Fprintf(w, "%s", plaintext)
	if err != nil {
		return err
	}
	return nil
}
