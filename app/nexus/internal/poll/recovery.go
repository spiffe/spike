package poll

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/cloudflare/circl/group"
	"github.com/cloudflare/circl/secretsharing"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
	"github.com/spiffe/spike-sdk-go/api/entity/v1/reqres"
	network "github.com/spiffe/spike-sdk-go/net"
	"github.com/spiffe/spike/app/nexus/internal/env"
	state "github.com/spiffe/spike/app/nexus/internal/state/base"
	"github.com/spiffe/spike/internal/auth"
	"github.com/spiffe/spike/internal/config"
	"github.com/spiffe/spike/internal/log"
	"github.com/spiffe/spike/internal/net"
	"github.com/spiffe/spike/pkg/crypto"
	"net/url"
	"os"
	"path"
	"sync"
	"time"
)

var rootKey []byte
var rootKeyMu sync.RWMutex

func shardUrl(keeperApiRoot string) string {
	u, err := url.JoinPath(keeperApiRoot, string(net.SpikeKeeperUrlShard))
	if err != nil {
		log.Log().Warn(
			"tick", "msg", "Failed to join path", "url", keeperApiRoot,
		)
		return ""
	}
	return u
}

func shardResponse(
	source *workloadapi.X509Source, u string, keeperId string,
) []byte {
	shardRequest := reqres.ShardRequest{}
	md, err := json.Marshal(shardRequest)
	if err != nil {
		log.Log().Warn("tick",
			"msg", "Failed to marshal request",
			"err", err, "keeper_id", keeperId)
		return []byte{}
	}

	client, err := network.CreateMtlsClientWithPredicate(
		source, auth.IsKeeper,
	)

	if err != nil {
		log.Log().Warn("tick",
			"msg", "Failed to create mTLS client",
			"err", err)
		return []byte{}
	}

	data, err := net.Post(client, u, md)
	if err != nil {
		log.Log().Warn("tick", "msg",
			"Failed to post",
			"err", err, "keeper_id", keeperId)
	}

	if len(data) == 0 {
		log.Log().Info("tick", "msg", "No data")
		return []byte{}
	}

	return data
}

func unmarshalShardResponse(data []byte) *reqres.ShardResponse {
	var res reqres.ShardResponse
	err := json.Unmarshal(data, &res)
	if err != nil {
		log.Log().Info("tick", "msg",
			"Failed to unmarshal response", "err", err)
		return nil
	}
	return &res
}

func rawShards(successfulKeeperShards map[string]string) [][]byte {
	ss := make([][]byte, 0)

	for keeperId, shard := range successfulKeeperShards {
		decodedShard, err := base64.StdEncoding.DecodeString(shard)
		if err != nil {
			log.Log().Warn("tick",
				"msg", "Failed to decode shard from base64",
				"err", err, "keeper_id", keeperId)
			return [][]byte{{}}
		}
		ss = append(ss, decodedShard)
	}

	return ss
}

func recoverRootKey(ss [][]byte) []byte {
	g := group.P256
	firstShard := ss[0]
	secondShard := ss[1]
	firstShare := secretsharing.Share{
		ID:    g.NewScalar(),
		Value: g.NewScalar(),
	}
	firstShare.ID.SetUint64(1)
	err := firstShare.Value.UnmarshalBinary(firstShard)
	if err != nil {
		log.FatalLn("Failed to unmarshal share: " + err.Error())
	}
	secondShare := secretsharing.Share{
		ID:    g.NewScalar(),
		Value: g.NewScalar(),
	}
	secondShare.ID.SetUint64(2)
	err = secondShare.Value.UnmarshalBinary(secondShard)
	if err != nil {
		log.FatalLn("Failed to unmarshal share: " + err.Error())
	}

	var shares []secretsharing.Share
	shares = append(shares, firstShare)
	shares = append(shares, secondShare)

	reconstructed, err := secretsharing.Recover(1, shares)
	if err != nil {
		log.FatalLn("Failed to recover: " + err.Error())
	}

	// TODO: check for errors.

	if reconstructed == nil {
		log.FatalLn("Failed to reconstruct the root key")
		return []byte{}
	}

	binaryRec, err := reconstructed.MarshalBinary()
	if err != nil {
		log.FatalLn("Failed to marshal: " + err.Error())
		return []byte{}
	}

	// TODO: check size 32bytes.

	return binaryRec
}

func iterateKeepersToRecover(
	source *workloadapi.X509Source,
	successfulKeeperShards map[string]string,
) bool {
	for keeperId, keeperApiRoot := range env.Keepers() {
		log.Log().Info("keeper", "id", keeperId, "url", keeperApiRoot)

		u := shardUrl(keeperApiRoot)
		if u == "" {
			continue
		}

		data := shardResponse(source, u, keeperId)
		if len(data) == 0 {
			continue
		}

		res := unmarshalShardResponse(data)
		if res == nil {
			continue
		}

		shard := res.Shard
		if len(shard) == 0 {
			log.Log().Info("tick", "msg", "No shard")
			continue
		}

		successfulKeeperShards[keeperId] = shard
		if len(successfulKeeperShards) != 2 {
			continue
		}

		// TODO: combine shards to create a root key.

		ss := rawShards(successfulKeeperShards)
		if len(ss) != 2 {
			continue
		}

		binaryRec := recoverRootKey(ss)
		encoded := hex.EncodeToString(binaryRec)
		state.Initialize(encoded)

		// TODO: all async persist operations will be sync.
		// also create an ADR for that.

		fmt.Println(">>>>>>>>>>>>>>>> RECOVERY 0001")
		rootKeyMu.Lock()
		rootKey = binaryRec
		rootKeyMu.Unlock()

		// System initialized: Exit infinite loop.
		return true
	}

	return false
}

// recoverUsingKeeperShards iterates through keepers until you get two shards.
//
// Any 400 and 5xx response that a SPIKE Keeper gives is likely temporary.
// We should keep trying until we get a 200 or 404 response.
func recoverUsingKeeperShards(source *workloadapi.X509Source) {
	successfulKeeperShards := make(map[string]string)

	for {
		exit := iterateKeepersToRecover(source, successfulKeeperShards)
		if exit {
			return
		}
		log.Log().Info("tick", "msg", "Waiting for keepers to respond")
		time.Sleep(5 * time.Second)
	}
}

//func recoveryInfoExists() bool {
//	recoveryInfo := persist.ReadRecoveryInfo()
//	return recoveryInfo != nil
//}

// TODO: name misleadiong; we dont' persist.
func mustPersistRecoveryInfo(rk string) []secretsharing.Share {
	decodedRootKey, err := hex.DecodeString(rk)
	if err != nil {
		log.FatalLn("Tick: failed to decode root key: " + err.Error())
	}
	rootSecret, rootShares := computeShares(decodedRootKey)
	sanityCheck(rootSecret, rootShares)

	//share1, err := rootShares[0].Value.MarshalBinary()
	//if err != nil {
	//	log.FatalLn("Tick: failed to marshal share: " + err.Error())
	//}
	//share2, err := rootShares[1].Value.MarshalBinary()
	//if err != nil {
	//	log.FatalLn("Tick: failed to marshal share: " + err.Error())
	//}

	// Save recovery information.
	fmt.Println(">>>>>>>>>>>>>>>> RECOVERY 0002")
	rootKeyMu.Lock()
	rootKey = decodedRootKey
	rootKeyMu.Unlock()

	//persist.AsyncPersistRecoveryInfo(store.KeyRecoveryData{
	//	RootKey:     decodedRootKey,
	//	MinShards:   2,
	//	TotalShards: 3,
	//	Shards:      [][]byte{share1, share2},
	//	CreatedTime: time.Time{},
	//	UpdatedTime: time.Time{},
	//})

	return rootShares
}

func shardContributionResponse(
	keeperId string, keepers map[string]string, u string,
	rootShares []secretsharing.Share, source *workloadapi.X509Source,
) []byte {
	fmt.Println("shardContributionResponse: before creating mTLS client")
	client, err := network.CreateMtlsClientWithPredicate(
		source, auth.IsKeeper,
	)
	fmt.Println("shardContributionResponse: after creating mTLS client")

	if err != nil {
		log.Log().Warn("tick",
			"msg", "Failed to create mTLS client",
			"err", err)
		return []byte{}
	}

	share := findShare(keeperId, keepers, rootShares)

	contribution, err := share.Value.MarshalBinary()
	if err != nil {
		log.Log().Warn("tick",
			"msg", "Failed to marshal share",
			"err", err, "keeper_id", keeperId)
		return []byte{}
	}

	scr := reqres.ShardContributionRequest{
		KeeperId: keeperId,
		Shard:    base64.StdEncoding.EncodeToString(contribution),
	}
	md, err := json.Marshal(scr)
	if err != nil {
		log.Log().Warn("tick",
			"msg", "Failed to marshal request",
			"err", err, "keeper_id", keeperId)
		return []byte{}
	}

	fmt.Println("shardContributionResponse: before post")

	data, err := net.Post(client, u, md)
	if err != nil {
		log.Log().Warn("tick", "msg",
			"Failed to post",
			"err", err, "keeper_id", keeperId)
	}

	fmt.Println("shardContributionResponse: after post 001")

	if len(data) == 0 {
		log.Log().Info("tick", "msg", "No data")
		return []byte{}
	}

	fmt.Println("shardContributionResponse: after post 002")
	return data
}

func iterateKeepersToBootstrap(
	keepers map[string]string, rootShares []secretsharing.Share,
	successfulKeepers map[string]bool, source *workloadapi.X509Source,
) bool {
	fmt.Println(">>>> in iterateKeepersToBootstrap")

	for keeperId, keeperApiRoot := range keepers {
		fmt.Println(">>>> in for: keeperId", keeperId)

		u, err := url.JoinPath(keeperApiRoot, string(net.SpikeKeeperUrlContribute))

		if err != nil {
			log.Log().Warn(
				"tick", "msg", "Failed to join path", "url", keeperApiRoot,
			)
			continue
		}

		data := shardContributionResponse(
			keeperId, keepers, u, rootShares, source,
		)
		if len(data) == 0 {
			fmt.Println(">>>> in if: len(data) == 0")
			continue
		}

		fmt.Println(">>>> in if: len(data) > 0")

		var res reqres.ShardContributionResponse
		err = json.Unmarshal(data, &res)
		if err != nil {
			log.Log().Info("tick", "msg",
				"Failed to unmarshal response", "err", err)
			continue
		}

		fmt.Println(">>>> in if: unmarshal success")

		successfulKeepers[keeperId] = true
		log.Log().Info("tick", "msg", "Success", "keeper_id", keeperId)

		if len(successfulKeepers) == 3 {
			log.Log().Info("tick", "msg", "All keepers initialized")

			tombstone := path.Join(
				config.SpikeNexusDataFolder(), "bootstrap.tombstone",
			)

			// Create the tombstone file to mark SPIKE Nexus as bootstrapped.
			err = os.WriteFile(
				tombstone,
				[]byte("spike.nexus.bootstrapped=true"), 0644,
			)
			if err != nil {
				// Although the tombstone file is just a marker, it's still important.
				// If SPIKE Nexus cannot create the tombstone file, or if someone
				// deletes it, this can slightly change system operations.
				// To be on the safe side, we let SPIKE Nexus crash because not being
				// able to write to the data volume (where the tombstone file would be)
				// can be a precursor of other problems that can affect the reliability
				// of the backing store.
				log.FatalLn("Tick: failed to create tombstone file: " + err.Error())
			}

			log.Log().Info("tick", "msg", "Tombstone file created successfully")

			return true
		}
	}

	return false
}

func sendShards(source *workloadapi.X509Source) {
	log.Log().Info("sendShards", "msg", "Will send shards to keepers")
	// TODO: this should be configurable.
	ticker := time.NewTicker(13 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			log.Log().Info("sendShards", "msg", "Sending shards to keepers")

			keepers := env.Keepers()
			if len(keepers) < 3 {
				log.FatalLn("sendShards: not enough keepers")
			}

			for keeperId, keeperApiRoot := range keepers {
				u, err := url.JoinPath(
					keeperApiRoot, string(net.SpikeKeeperUrlContribute),
				)

				if err != nil {
					log.Log().Warn(
						"sendShards", "msg", "Failed to join path", "url", keeperApiRoot,
					)
					continue
				}

				client, err := network.CreateMtlsClientWithPredicate(
					source, auth.IsKeeper,
				)

				if err != nil {
					log.Log().Warn("sendShards",
						"msg", "Failed to create mTLS client",
						"err", err)
					continue
				}

				// TODO: only root key is enough for recovery info; we
				// can always compute shares. remove other data from the struct.
				// recoveryInfo := persist.ReadRecoveryInfo()

				rootKeyMu.RLock()
				if rootKey == nil {
					fmt.Println("sendShards: recoveryInfo is nil; moving on...")
					rootKeyMu.RUnlock()
					continue
				}

				rootSecret, rootShares := computeShares(rootKey)
				rootKeyMu.RUnlock()

				sanityCheck(rootSecret, rootShares)

				share := findShare(keeperId, keepers, rootShares)

				contribution, err := share.Value.MarshalBinary()
				if err != nil {
					log.Log().Warn("sendShards",
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
					log.Log().Warn("sendShards",
						"msg", "Failed to marshal request",
						"err", err, "keeper_id", keeperId)
					continue
				}

				_, err = net.Post(client, u, md)
				if err != nil {
					log.Log().Warn("sendShards", "msg",
						"Failed to post",
						"err", err, "keeper_id", keeperId)
				}
			}
		}
	}
}

func bootstrapBackingStore(source *workloadapi.X509Source) {
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
		log.FatalLn("Tick: failed to create root key: " + err.Error())
	}

	// Initialize the backend store before sending shards to the keepers.
	// SPIKE Keepers are our backup system, and they are not critical for system
	// operations. Initializing early allows SPIKE Nexus to serve before
	// keepers are hydrated.
	state.Initialize(rk)
	log.Log().Info("tick", "msg", "Initialized the backing store")
	// Compute Shamir shares out of the root key.
	rootShares := mustPersistRecoveryInfo(rk)

	successfulKeepers := make(map[string]bool)
	keepers := env.Keepers()
	if len(keepers) < 3 {
		log.FatalLn("Tick: not enough keepers")
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
