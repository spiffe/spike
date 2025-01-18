package recovery

import (
	"encoding/base64"
	"encoding/json"
	"net/url"
	"time"

	"github.com/spiffe/go-spiffe/v2/workloadapi"
	"github.com/spiffe/spike-sdk-go/api/entity/v1/reqres"
	network "github.com/spiffe/spike-sdk-go/net"

	"github.com/spiffe/spike/app/nexus/internal/env"
	state "github.com/spiffe/spike/app/nexus/internal/state/base"
	"github.com/spiffe/spike/internal/auth"
	"github.com/spiffe/spike/internal/log"
	"github.com/spiffe/spike/internal/net"
	"github.com/spiffe/spike/pkg/crypto"
)

// RecoverBackingStoreUsingKeeperShards iterates through keepers until
// you get two shards.
//
// Any 400 and 5xx response that a SPIKE Keeper gives is likely temporary.
// We should keep trying until we get a 200 or 404 response.
//
// This function attempts to recover the backing store by collecting shards
// from keeper nodes. It continuously polls the keepers until enough valid
// shards are collected to reconstruct the backing store. The function blocks\
// until recovery is successful.
//
// The function maintains a map of successfully recovered shards from each
// keeper to avoid duplicate processing. It waits for 5 seconds between retry
// attempts if recovery is unsuccessful.
//
// Parameters:
//   - source: An X509Source used for authenticating with keeper nodes
func RecoverBackingStoreUsingKeeperShards(source *workloadapi.X509Source) {
	successfulKeeperShards := make(map[string]string)

	for {
		recoverySuccessful := iterateKeepersAndTryRecovery(
			source, successfulKeeperShards,
		)
		if recoverySuccessful {
			return
		}
		log.Log().Info("tick", "msg", "Waiting for keepers to respond")
		time.Sleep(5 * time.Second)
	}
}

// SendShardsPeriodically distributes key shards to configured keeper nodes at
// regular intervals. It creates new shards from the current root key and sends
// them to each keeper using mTLS authentication. The function runs indefinitely
// until stopped.
//
// The function sends shards every 13 seconds (configurable in future). It
// requires a minimum of 3 keepers to be configured. If any operation fails for
// a keeper (URL creation, mTLS setup, marshaling, or network request), it logs
// a warning and continues with the next keeper.
//
// Parameters:
//   - source: An X509Source used for creating mTLS connections to keepers
func SendShardsPeriodically(source *workloadapi.X509Source) {
	const fName = "SendShardsPeriodically"

	log.Log().Info(fName, "msg", "Will send shards to keepers")

	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		log.Log().Info(fName, "msg", "Sending shards to keepers")

		keepers := env.Keepers()
		if len(keepers) < 3 {
			log.FatalLn(fName + ": not enough keepers")
		}

		for keeperId, keeperApiRoot := range keepers {
			u, err := url.JoinPath(
				keeperApiRoot, string(net.SpikeKeeperUrlContribute),
			)

			if err != nil {
				log.Log().Warn(
					fName, "msg", "Failed to join path", "url", keeperApiRoot,
				)
				continue
			}

			client, err := network.CreateMtlsClientWithPredicate(
				source, auth.IsKeeper,
			)

			if err != nil {
				log.Log().Warn(fName,
					"msg", "Failed to create mTLS client",
					"err", err)
				continue
			}

			rootKeyMu.RLock()
			if rootKey == nil {
				log.Log().Info(fName, "msg", "rootKey is nil; moving on...")
				rootKeyMu.RUnlock()
				continue
			}

			rootSecret, rootShares := computeShares(rootKey)
			rootKeyMu.RUnlock()

			sanityCheck(rootSecret, rootShares)

			share := findShare(keeperId, keepers, rootShares)

			contribution, err := share.Value.MarshalBinary()
			if err != nil {
				log.Log().Warn(fName,
					"msg", "Failed to marshal share",
					"err", err, "keeper_id", keeperId)
				continue
			}

			scr := reqres.ShardContributionRequest{
				KeeperId: keeperId,
				Shard:    base64.StdEncoding.EncodeToString(contribution),
			}
			md, err := json.Marshal(scr)
			if err != nil {
				log.Log().Warn(fName,
					"msg", "Failed to marshal request",
					"err", err, "keeper_id", keeperId)
				continue
			}

			_, err = net.Post(client, u, md)
			if err != nil {
				log.Log().Warn(fName, "msg",
					"Failed to post",
					"err", err, "keeper_id", keeperId)
			}
		}
	}
}

// BootstrapBackingStoreWithNewRootKey initializes the backing store with a new
// root key if it hasn't been bootstrapped already. It generates a new AES-256
// root key, initializes the state with this key, and distributes key shards
// to all configured keepers.
//
// The function requires a minimum of 3 keepers to be configured. It
// continuously attempts to distribute shards to all keepers until successful,
// waiting 5 seconds between retry attempts. The backing store is initialized
// before keeper distribution to allow immediate operation.
//
// Parameters:
//   - source: An X509Source used for authenticating with keeper nodes
//
// The function will fatal if:
//   - Root key creation fails
//   - Fewer than 3 keepers are configured
func BootstrapBackingStoreWithNewRootKey(source *workloadapi.X509Source) {
	log.Log().Info("tick", "msg",
		"Tombstone file does not exist. Bootstrapping SPIKE Nexus...")

	rootKeyMu.RLock()

	if rootKey != nil {
		log.Log().Info("tick", "msg",
			"Recovery info found. Backing store already bootstrapped.",
		)
		rootKeyMu.RUnlock()
		return
	}
	rootKeyMu.RUnlock()

	// Create the root key and create shards out of the root key.
	rk, err := crypto.Aes256Seed()
	if err != nil {
		log.FatalLn("Bootstrap: failed to create root key: " + err.Error())
	}

	// Initialize the backend store before sending shards to the keepers.
	// SPIKE Keepers are our backup system, and they are not critical for system
	// operations. Initializing early allows SPIKE Nexus to serve before
	// keepers are hydrated.
	state.Initialize(rk)
	log.Log().Info("tick", "msg", "Initialized the backing store")

	// Compute Shamir shares out of the root key.
	rootShares := mustUpdateRecoveryInfo(rk)

	successfulKeepers := make(map[string]bool)
	keepers := env.Keepers()
	if len(keepers) < 3 {
		log.FatalLn("Bootstrap: not enough keepers")
	}

	for {
		// Ensure to get a success response from ALL keepers eventually.
		exit := iterateKeepersToBootstrap(
			keepers, rootShares, successfulKeepers, source,
		)
		if exit {
			return
		}

		log.Log().Info("tick", "msg", "Waiting for keepers to initialize")
		time.Sleep(5 * time.Second)
	}
}
