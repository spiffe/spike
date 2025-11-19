package net

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"io"
	"time"

	"github.com/spiffe/go-spiffe/v2/workloadapi"
	spike "github.com/spiffe/spike-sdk-go/api"
	"github.com/spiffe/spike-sdk-go/config/env"
	sdkErrors "github.com/spiffe/spike-sdk-go/errors"
	"github.com/spiffe/spike-sdk-go/log"
	tls "github.com/spiffe/spike-sdk-go/net"
	"github.com/spiffe/spike-sdk-go/retry"
	svid "github.com/spiffe/spike-sdk-go/spiffeid"

	"github.com/spiffe/spike/app/bootstrap/internal/state"
)

// BroadcastKeepers distributes root key shares to all configured SPIKE Keeper
// instances. It iterates through each keeper ID from the environment
// configuration and sends the corresponding keeper share using the provided
// API. The function retries indefinitely until each share is successfully
// delivered. If a keeper fails to receive its share, the function logs a
// warning and retries. The function terminates the application if the retry
// mechanism fails unexpectedly.
func BroadcastKeepers(ctx context.Context, api *spike.API) {
	const fName = "BroadcastKeepers"

	// TODO: nil check for all ctx context.Context args.

	// RootShares() generates the root key and splits it into shares.
	// It enforces single-call semantics and will terminate if called again.
	rs := state.RootShares()

	for keeperID := range env.KeepersVal() {
		keeperShare := state.KeeperShare(rs, keeperID)

		log.Log().Info(fName, "message", "iterating", "keeper ID", keeperID)

		_, err := retry.Forever(ctx, func() (bool, error) {
			log.Log().Info(fName, "message", "retry:"+time.Now().String())
			err := api.Contribute(keeperShare, keeperID)
			if err != nil {
				failErr := sdkErrors.ErrPostFailed.Wrap(err)
				failErr.Msg = "failed to send shard: will retry"
				log.WarnErr(fName, *failErr)
				return false, failErr
			}

			log.Log().Info(fName, "message", "shard sent successfully")
			return true, nil
		})

		// This should never happen since the above loop retries forever:
		if err != nil {
			log.FatalLn(fName, "message", "initialization failed", "err", err)
		}
	}
}

// VerifyInitialization confirms that the SPIKE Nexus initialization was
// successful by performing an end-to-end encryption test. The function
// generates a random 32-byte value, encrypts it using AES-GCM with the root
// key, and sends the plaintext along with the nonce and ciphertext to SPIKE
// Nexus for verification. The function retries indefinitely until the
// verification succeeds. It terminates the application if any cryptographic
// operations fail.
func VerifyInitialization(ctx context.Context, api *spike.API) {
	const fName = "VerifyInitialization"

	// Generate random text for verification
	randomBytes := make([]byte, 32)
	_, err := rand.Read(randomBytes)
	if err != nil {
		log.FatalLn(
			fName,
			"message", "failed to generate random text",
			"err", err.Error(),
		)
		return
	}
	randomText := hex.EncodeToString(randomBytes)
	log.Log().Info(
		fName,
		"message", "generated random verification text",
		"length", len(randomText),
	)

	// Encrypt the random text with the root key
	rootKey := state.RootKey()
	block, err := aes.NewCipher(rootKey[:])
	if err != nil {
		log.FatalLn(
			fName, "message",
			"failed to create cipher",
			"err", err.Error(),
		)
		return
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		log.FatalLn(
			fName,
			"message", "failed to create GCM",
			"err", err.Error(),
		)
		return
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		log.FatalLn(
			fName,
			"message", "failed to generate nonce",
			"err", err.Error(),
		)
		return
	}

	ciphertext := gcm.Seal(nil, nonce, []byte(randomText), nil)
	log.Log().Info(
		fName,
		"message", "encrypted verification text",
		"nonce_len", len(nonce),
		"ciphertext_len", len(ciphertext),
	)

	_, _ = retry.Forever(ctx, func() (bool, error) {
		log.Log().Info(
			fName,
			"message", "retry:"+time.Now().String(),
		)
		err := api.Verify(randomText, nonce, ciphertext)
		if err != nil {
			log.Log().Warn(
				fName,
				"message", "failed to verify signature",
				"err", err.Error(),
			)
			return false, err
		}
		return true, nil
	})
}

// AcquireSource obtains and validates an X.509 SVID source with a SPIKE
// Bootstrap SPIFFE ID. The function retrieves the X.509 SVID from the SPIFFE
// Workload API and verifies that the SPIFFE ID matches the expected ID pattern.
// If the SVID cannot be obtained or does not have the required bootstrap
// SPIFFE ID, the function terminates the application. This function is used
// to ensure that only authorized SPIKE Bootstrap workloads can perform
// initialization operations.
func AcquireSource() *workloadapi.X509Source {
	const fName = "AcquireSource"

	src := tls.Source()
	sv, err := src.GetX509SVID()
	if err != nil {
		log.FatalLn(fName,
			"message", "Failed to get X.509 SVID",
			"err", err.Error())
		log.FatalLn(fName, "message", "Failed to acquire SVID")
		return nil
	}
	if !svid.IsBootstrap(sv.ID.String()) {
		log.Log().Error(
			fName,
			"message", "you need a bootstrap SPIFFE ID to use this command",
		)
		log.FatalLn(fName, "message", "command not authorized")
		return nil
	}

	if src == nil {
		log.FatalLn(fName, "message", "failed to acquire SVID")
		return nil
	}

	return src
}
