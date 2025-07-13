package secret

import (
	"crypto/rand"
	"fmt"
	"io"
	"net/http"

	"github.com/spiffe/spike-sdk-go/api/errors"

	"github.com/spiffe/spike/app/nexus/internal/state/persist"
	"github.com/spiffe/spike/internal/log"
	"github.com/spiffe/spike/internal/net"
)

func RouteEncrypt(
	w http.ResponseWriter, r *http.Request, audit *log.AuditEntry,
) error {
	const fName = "routeEncrypt"
	log.AuditRequest(fName, r, audit, log.AuditCreate)

	requestBody := net.ReadRequestBody(w, r)
	if requestBody == nil {
		return errors.ErrReadFailure
	}

	c := persist.Backend().GetCipher()
	nonce := make([]byte, c.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return fmt.Errorf("failed to generate nonce: %w", err)
	}
	log.Log().Info(fName, "msg", fmt.Sprintf("Encrypt %d %d", len(nonce), len(requestBody)))
	ciphertext := c.Seal(nil, nonce, requestBody, nil)
	log.Log().Info(fName, "msg", fmt.Sprintf("len after %d %d", len(nonce), len(ciphertext)))
	v := byte('1')
	w.Write([]byte{v})
	w.Write(nonce)
	w.Write(ciphertext)
	return nil
}
