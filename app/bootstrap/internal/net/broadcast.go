//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package net

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"

	"github.com/spiffe/go-spiffe/v2/workloadapi"
	spike "github.com/spiffe/spike-sdk-go/api"
	"github.com/spiffe/spike-sdk-go/config/env"
	"github.com/spiffe/spike-sdk-go/crypto"
	sdkErrors "github.com/spiffe/spike-sdk-go/errors"
	"github.com/spiffe/spike-sdk-go/log"
	"github.com/spiffe/spike-sdk-go/retry"
	"github.com/spiffe/spike-sdk-go/spiffe"
	svid "github.com/spiffe/spike-sdk-go/spiffeid"
	"github.com/spiffe/spike-sdk-go/validation"

	"github.com/spiffe/spike/app/bootstrap/internal/state"
)

// BroadcastKeepers distributes root key shares to all configured SPIKE
// Keeper instances. It iterates through each keeper ID from the
// environment configuration and sends the corresponding keeper share
// using the provided API. The function retries with exponential
// backoff using a configurable timeout and a maximum number of retry
// attempts per keeper. If a keeper cannot be reached within the
// timeout, the function terminates with a clear error message,
// allowing the operator to fix the issue and rerun bootstrap.
//
// This fail-fast approach is appropriate for bootstrap because:
//   - Bootstrap is a day-zero operation with an active operator watching
//   - Bootstrap is idempotent and safe to rerun after fixing issues
//   - Clear error messages help operators quickly identify and fix
//     problems
//   - Blocking forever obscures problems rather than surfacing them
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//   - api: SPIKE API client for communicating with keepers
func BroadcastKeepers(ctx context.Context, api *spike.API) {
	const fName = "BroadcastKeepers"

	validation.NonNilContextOrDie(ctx, fName)

	// Ensures the number of keepers matches the Shamir shares required.
	keepers := env.KeepersVal()
	expectedShares := env.ShamirSharesVal()

	if len(keepers) != expectedShares {
		failErr := sdkErrors.ErrShamirNotEnoughKeepers.Clone()
		failErr.Msg = fmt.Sprintf(
			"keeper count mismatch: SPIKE_NEXUS_SHAMIR_SHARES=%d "+
				"but %d keepers configured in SPIKE_NEXUS_KEEPER_PEERS; "+
				"these values must match",
			expectedShares, len(keepers),
		)
		log.FatalErr(fName, *failErr)
		return
	}

	state.LockRootKeySeed()
	defer state.UnlockRootKeySeed()
	// RootShares() generates the root key and splits it into shares.
	// It enforces single-call semantics and will terminate if called again.
	rks := state.RootKeySeedNoLock()
	rs := crypto.RootShares(rks)

	timeout := env.BootstrapKeeperTimeoutVal()
	maxRetries := env.BootstrapKeeperMaxRetriesVal()

	for keeperID, keeperURL := range env.KeepersVal() {
		err := broadcastToKeeper(
			ctx, api, rs, keeperID, keeperURL, timeout, maxRetries,
		)
		if err != nil {
			failErr := sdkErrors.ErrBootstrapKeeperUnreachable.Clone()
			failErr.Msg = fmt.Sprintf(
				"failed to reach keeper %s at %s after %d attempts; "+
					"ensure all keepers are running and rerun bootstrap",
				keeperID, keeperURL, maxRetries,
			)
			log.FatalErr(fName, *failErr)
			return
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
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//   - api: SPIKE API client for verification requests
func VerifyInitialization(ctx context.Context, api *spike.API) {
	const fName = "VerifyInitialization"

	validation.NonNilContextOrDie(ctx, fName)

	// Generate random text for verification
	randomBytes := make([]byte, 32)
	_, err := rand.Read(randomBytes)
	if err != nil {
		failErr := sdkErrors.ErrCryptoRandomGenerationFailed.Wrap(err)
		log.FatalErr(fName, *failErr)
		return
	}
	randomText := hex.EncodeToString(randomBytes)

	state.LockRootKeySeed()
	// Encrypt the random text with the root key
	rootKey := state.RootKeySeed()
	defer state.UnlockRootKeySeed()
	block, aesErr := aes.NewCipher(rootKey[:])
	if aesErr != nil {
		failErr := sdkErrors.ErrCryptoFailedToCreateCipher.Wrap(aesErr)
		log.FatalErr(fName, *failErr)
		return
	}

	gcm, gcmErr := cipher.NewGCM(block)
	if gcmErr != nil {
		failErr := sdkErrors.ErrCryptoFailedToCreateGCM.Wrap(gcmErr)
		log.FatalErr(fName, *failErr)
		return
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, nonceErr := io.ReadFull(rand.Reader, nonce); nonceErr != nil {
		failErr := sdkErrors.ErrCryptoFailedToReadNonce.Wrap(nonceErr)
		log.FatalErr(fName, *failErr)
		return
	}

	ciphertext := gcm.Seal(nil, nonce, []byte(randomText), nil)

	// At this point we talk to SPIKE Nexus, and our expectation is SPIKE Nexus
	// is about to become healthy. So, we retry with a reasonable timeout and
	// give up if we cannot verify the initialization in a timely manner.

	_, retryErr := retry.Do(ctx, func() (bool, *sdkErrors.SDKError) {
		verifyErr := api.Verify(ctx, randomText, nonce, ciphertext)
		if verifyErr != nil {
			failErr := sdkErrors.ErrCryptoCipherVerificationFailed.Wrap(verifyErr)
			failErr.Msg = "failed to verify initialization: will retry"
			log.WarnErr(fName, *failErr)
			return false, failErr
		}
		return true, nil
	}, retry.WithBackOffOptions(
		retry.WithMaxElapsedTime(env.BootstrapInitVerificationTimeoutVal())),
	)

	if retryErr != nil {
		failErr := sdkErrors.ErrCryptoCipherVerificationFailed.Wrap(retryErr)
		failErr.Msg = "failed to verify initialization within timeout"
		log.FatalErr(fName, *failErr)
	}
}

// AcquireSource obtains and validates an X.509 SVID source with a SPIKE
// Bootstrap SPIFFE ID. The function retrieves the X.509 SVID from the SPIFFE
// Workload API and verifies that the SPIFFE ID matches the expected ID pattern.
// If the SVID cannot be obtained or does not have the required bootstrap
// SPIFFE ID, the function terminates the application. This function is used
// to ensure that only authorized SPIKE Bootstrap workloads can perform
// initialization operations.
//
// Returns:
//   - *workloadapi.X509Source: The validated X.509 SVID source, or nil if
//     acquisition fails (the function terminates the application on failure)
func AcquireSource() *workloadapi.X509Source {
	const fName = "AcquireSource"

	ctx, cancel := context.WithTimeout(
		context.Background(),
		env.SPIFFESourceTimeoutVal(),
	)
	defer cancel()

	src, spiffeID, err := spiffe.Source(ctx, spiffe.EndpointSocket())
	if err != nil {
		log.FatalErr(fName, *err)
		return nil
	}

	if !svid.IsBootstrap(spiffeID) {
		failErr := *sdkErrors.ErrAccessUnauthorized.Clone()
		failErr.Msg = "bootstrap SPIFFE ID required"
		log.FatalErr(fName, failErr)
		return nil
	}

	return src
}
