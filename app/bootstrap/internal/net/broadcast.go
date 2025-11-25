package net

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"io"

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

		log.Log().Info(fName, "message", "iterating", "keeper_id", keeperID)
		_, err := retry.Forever(ctx, func() (bool, *sdkErrors.SDKError) {
			log.Log().Info(fName, "message", "retrying", "keeper_id", keeperID)

			err := api.Contribute(keeperShare, keeperID)
			if err != nil {
				failErr := sdkErrors.ErrAPIPostFailed.Wrap(err)
				failErr.Msg = "failed to send shard: will retry"
				log.WarnErr(fName, *failErr)
				return false, failErr
			}

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
		failErr := sdkErrors.ErrCryptoRandomGenerationFailed.Wrap(err)
		log.FatalErr(fName, *failErr)
		return
	}
	randomText := hex.EncodeToString(randomBytes)

	// Encrypt the random text with the root key
	rootKey := state.RootKey()
	block, err := aes.NewCipher(rootKey[:])
	if err != nil {
		failErr := sdkErrors.ErrCryptoFailedToCreateCipher.Wrap(err)
		log.FatalErr(fName, *failErr)
		return
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		failErr := sdkErrors.ErrCryptoFailedToCreateGCM.Wrap(err)
		log.FatalErr(fName, *failErr)
		return
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		failErr := sdkErrors.ErrCryptoFailedToReadNonce.Wrap(err)
		log.FatalErr(fName, *failErr)
		return
	}

	ciphertext := gcm.Seal(nil, nonce, []byte(randomText), nil)

	_, _ = retry.Forever(ctx, func() (bool, *sdkErrors.SDKError) {
		err := api.Verify(randomText, nonce, ciphertext)
		if err != nil {
			failErr := sdkErrors.ErrCryptoCipherVerificationFailed.Wrap(err)
			failErr.Msg = "failed to verify initialization: will retry"
			log.WarnErr(fName, *failErr)
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
		failErr := sdkErrors.ErrSPIFFEFailedToExtractX509SVID.Wrap(err)
		log.FatalErr(fName, *failErr)
		return nil
	}
	if !svid.IsBootstrap(sv.ID.String()) {
		failErr := *sdkErrors.ErrAccessUnauthorized.Clone()
		failErr.Msg = "bootstrap SPIFFE ID required"
		log.FatalErr(fName, failErr)
		return nil
	}

	if src == nil {
		failErr := *sdkErrors.ErrSPIFFEFailedToExtractX509SVID.Clone()
		failErr.Msg = "failed to acquire X.509 SVID"
		log.FatalErr(fName, failErr)
		return nil
	}

	return src
}
