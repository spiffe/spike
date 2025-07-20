package secret

import (
	"crypto/rand"
	"fmt"
	"io"
	"net/http"

	"github.com/spiffe/spike-sdk-go/api/errors"
	"github.com/spiffe/spike-sdk-go/log"

	"github.com/spiffe/spike/app/nexus/internal/state/persist"
	journal "github.com/spiffe/spike/internal/log"
	"github.com/spiffe/spike/internal/net"
)

func RouteEncrypt(
	w http.ResponseWriter, r *http.Request, audit *journal.AuditEntry,
) error {
	const fName = "routeEncrypt"
	journal.AuditRequest(fName, r, audit, journal.AuditCreate)

	requestBody := net.ReadRequestBody(w, r)
	if requestBody == nil {
		return errors.ErrReadFailure
	}

	c := persist.Backend().GetCipher()
	nonce := make([]byte, c.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return fmt.Errorf("failed to generate nonce: %w", err)
	}
	log.Log().Info(fName, "message", fmt.Sprintf("Encrypt %d %d", len(nonce), len(requestBody)))
	ciphertext := c.Seal(nil, nonce, requestBody, nil)
	log.Log().Info(fName, "message", fmt.Sprintf("len after %d %d", len(nonce), len(ciphertext)))
	v := byte('1')
	_, err := w.Write([]byte{v})
	if err != nil {
		return err
	}
	_, err = w.Write(nonce)
	if err != nil {
		return err
	}
	_, err = w.Write(ciphertext)
	if err != nil {
		return err
	}
	return nil
}
