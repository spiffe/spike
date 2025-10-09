//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/fips140"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"time"

	"github.com/spiffe/spike-sdk-go/config/env"
	"github.com/spiffe/spike-sdk-go/log"
	"github.com/spiffe/spike-sdk-go/retry"
	"github.com/spiffe/spike-sdk-go/spiffe"
	svid "github.com/spiffe/spike-sdk-go/spiffeid"

	"github.com/spiffe/spike/app/bootstrap/internal/lifecycle"
	"github.com/spiffe/spike/app/bootstrap/internal/net"
	"github.com/spiffe/spike/app/bootstrap/internal/state"
	"github.com/spiffe/spike/app/bootstrap/internal/url"
	"github.com/spiffe/spike/internal/config"
)

// TODO: some folders are missing doc.go

func main() {
	const fName = "bootstrap.main"

	log.Log().Info(fName, "message",
		"Starting SPIKE bootstrap...",
		"version", config.BootstrapVersion,
	)

	init := flag.Bool("init", false, "Initialize the bootstrap module")
	flag.Parse()
	if !*init {
		fmt.Println("")
		fmt.Println("Usage: bootstrap -init")
		fmt.Println("")
		log.FatalLn(fName, "message", "Invalid command line arguments")
		return
	}

	skip := !lifecycle.ShouldBootstrap() // Kubernetes or bare-metal check.
	if skip {
		log.Log().Info(fName,
			"message", "Skipping bootstrap.",
		)
		fmt.Println("Bootstrap skipped. Check the logs for more information.")
		return
	}

	src := net.Source()
	defer spiffe.CloseSource(src)
	sv, err := src.GetX509SVID()
	if err != nil {
		log.FatalLn(fName,
			"message", "Failed to get X.509 SVID",
			"err", err.Error())
		log.FatalLn(fName, "message", "Failed to acquire SVID")
		return
	}

	if !svid.IsBootstrap(sv.ID.String()) {
		log.Log().Error(
			"Authenticate: You need a 'bootstrap' SPIFFE ID to use this command.",
		)
		log.FatalLn(fName, "message", "Command not authorized")
		return
	}

	log.Log().Info(
		fName, "FIPS 140.3 enabled", fips140.Enabled(),
	)

	log.Log().Info(
		fName, "message", "Sending shards to SPIKE Keeper instances...",
	)

	ctx := context.Background()

	for keeperID, keeperAPIRoot := range env.KeepersVal() {
		log.Log().Info(fName, "keeper ID", keeperID)

		_, err := retry.Do(ctx, func() (bool, error) {
			log.Log().Info(fName, "message", "retry:"+time.Now().String())

			err := net.Post(
				net.MTLSClient(src),
				url.KeeperContributeEndpoint(keeperAPIRoot),
				net.Payload(
					state.KeeperShare(
						state.RootShares(), keeperID),
					keeperID,
				),
				keeperID,
			)
			if err != nil {
				log.Log().Warn(fName, "message", "Failed to send shard. Will retry.")
				return false, err
			}

			log.Log().Info(fName, "message", "Shard sent successfully.")
			return true, nil
		},
			retry.WithBackOffOptions(
				retry.WithMaxInterval(60*time.Second), // TODO: to env vars.
				retry.WithMaxElapsedTime(0),           // Retry forever.
			),
		)

		// This should never happen since the above loop retries forever:
		if err != nil {
			log.FatalLn(fName, "message", "Initialization failed", "err", err)
		}
	}

	log.Log().Info(fName, "message", "Sent shards to SPIKE Keeper instances.")

	// Verify that SPIKE Nexus has been properly initialized by sending an
	// encrypted payload and verifying the hash of the decrypted plaintext.

	// Generate random text for verification
	randomBytes := make([]byte, 32)
	_, err = rand.Read(randomBytes)
	if err != nil {
		log.FatalLn(fName, "message",
			"Failed to generate random text", "err", err.Error())
		return
	}
	randomText := hex.EncodeToString(randomBytes)
	log.Log().Info(fName, "message",
		"Generated random verification text", "length", len(randomText))

	// Encrypt the random text with the root key
	rootKey := state.RootKey()
	block, err := aes.NewCipher(rootKey[:])
	if err != nil {
		log.FatalLn(fName, "message",
			"Failed to create cipher", "err", err.Error())
		return
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		log.FatalLn(fName, "message",
			"Failed to create GCM", "err", err.Error())
		return
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		log.FatalLn(fName, "message",
			"Failed to generate nonce", "err", err.Error())
		return
	}

	ciphertext := gcm.Seal(nil, nonce, []byte(randomText), nil)
	log.Log().Info(fName, "message",
		"Encrypted verification text",
		"nonce_len", len(nonce),
		"ciphertext_len", len(ciphertext))

	// Send verification request to SPIKE Nexus
	nexusAPIRoot := env.NexusAPIRootVal()
	verifyURL := url.NexusVerifyEndpoint(nexusAPIRoot)

	log.Log().Info(fName, "message",
		"Sending verification request to SPIKE Nexus", "url", verifyURL)

	nexusClient := net.MTLSClientForNexus(src)
	verifyPayload := net.VerifyPayload(nonce, ciphertext)

	responseBody, err := net.PostVerify(nexusClient, verifyURL, verifyPayload)
	if err != nil {
		log.FatalLn(fName, "message",
			"Failed to send verification request", "err", err.Error())
		return
	}

	// Parse the response
	var verifyResponse struct {
		Hash string `json:"hash"`
		Err  string `json:"err"`
	}
	if err := json.Unmarshal(responseBody, &verifyResponse); err != nil {
		log.FatalLn(fName, "message",
			"Failed to parse verification response", "err", err.Error())
		return
	}

	// Compute expected hash
	expectedHash := sha256.Sum256([]byte(randomText))
	expectedHashHex := hex.EncodeToString(expectedHash[:])

	// Verify the hash matches
	if verifyResponse.Hash != expectedHashHex {
		log.FatalLn(fName, "message",
			"Verification failed: hash mismatch",
			"expected", expectedHashHex,
			"received", verifyResponse.Hash)
		return
	}

	log.Log().Info(fName, "message",
		"SPIKE Nexus verification successful", "hash", verifyResponse.Hash)

	// Mark completion in Kubernetes
	if err := lifecycle.MarkBootstrapComplete(); err != nil {
		// Log but don't fail - bootstrap itself succeeded
		log.Log().Warn(fName, "message",
			"Could not mark bootstrap complete in ConfigMap", "err", err.Error())
	}

	fmt.Println("Bootstrap completed successfully!")
}
