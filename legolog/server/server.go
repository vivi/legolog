package legolog

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/huyuncong/MerkleSquare/core"
	libcrypto "github.com/huyuncong/MerkleSquare/lib/crypto"
	"github.com/huyuncong/MerkleSquare/lib/storage"

	"github.com/immesys/bw2/crypto"
)

type Server struct {
	PartitionServers      []*PartitionServer
	epoch                 uint64
	epochLock             *sync.RWMutex
	verificationEpoch     uint64
	verificationEpochLock *sync.RWMutex
	updateEpochDuration   time.Duration
	verifyEpochDuration   time.Duration
	stopper               chan struct{}
}

type PartitionServer struct {
	Partition core.LegoLogPartition
	Storage   storage.Storage

	LastPos     uint64
	LastPosLock *sync.RWMutex

	NeedToRollUp     bool
	NeedToRollUpLock *sync.Mutex

	PublishedPos    uint64
	PublishedDigest core.LegologDigest

	AppendLock *sync.Mutex
	Index      int
}

type ValueRecord struct {
	Position  uint64
	Signature []byte
	Value     []byte
}

func (s *Server) IncrementUpdateEpoch() error {

	s.epochLock.Lock()
	s.epoch++
	var wg sync.WaitGroup
	for _, partitionServer := range s.PartitionServers {
		wg.Add(1)
		go func(partitionServer *PartitionServer) {
			partitionServer.IncrementUpdateEpoch()
			wg.Done()
		}(partitionServer)
	}
	wg.Wait()
	s.epochLock.Unlock()
	return nil
}

// Should only be called from server's increment epoch
func (partitionServer *PartitionServer) IncrementUpdateEpoch() {
	partitionServer.LastPosLock.RLock()

	/* Always increment the update epoch, even if we don't have any new updates.
	if partitionServer.PublishedPos == partitionServer.LastPos {
		partitionServer.LastPosLock.RUnlock()
		return
	}
	*/
	partitionServer.PublishedPos = partitionServer.LastPos
	partitionServer.Partition.IncrementUpdateEpoch()
	partitionServer.PublishedDigest = *partitionServer.Partition.GetDigest()
	// fmt.Printf("Just set the digest for partition server %d with roots[0] as %s\n", partitionServer.Index, partitionServer.PublishedDigest.UpdateSetRoots[0])
	partitionServer.LastPosLock.RUnlock()

	partitionServer.NeedToRollUpLock.Lock()
	partitionServer.NeedToRollUp = true
	partitionServer.NeedToRollUpLock.Unlock()

	//Dump obsolete extension proofs.
	/* 	s.extensionProofCache = make(map[ExtensionProofKey][]byte)
	 */

}

func (s *Server) IncrementVerificationPeriod() error {
	s.epochLock.Lock()
	s.verificationEpoch += 1
	var wg sync.WaitGroup
	for i, partitionServer := range s.PartitionServers {
		wg.Add(1)
		go func(i int, partitionServer *PartitionServer) {
			partitionServer.IncrementVerificationPeriod()
			//fmt.Println(i, "pos is ", partitionServer.PublishedPos)
			wg.Done()

		}(i, partitionServer)
	}
	wg.Wait()
	s.epochLock.Unlock()
	return nil
}

// Should only be called from server's increment epoch
func (partitionServer *PartitionServer) IncrementVerificationPeriod() {
	partitionServer.NeedToRollUpLock.Lock()

	if !partitionServer.NeedToRollUp {
		partitionServer.NeedToRollUpLock.Unlock()
		return
	}
	partitionServer.Partition.IncrementVerificationPeriod()
	partitionServer.PublishedDigest = *partitionServer.Partition.GetDigest()
	// fmt.Printf("Just set the digest for partition server %d with roots[0] as %s\n", partitionServer.Index, partitionServer.PublishedDigest.UpdateSetRoots[0])
	partitionServer.NeedToRollUp = false
	partitionServer.NeedToRollUpLock.Unlock()

	//Dump obsolete extension proofs.
	/* 	s.extensionProofCache = make(map[ExtensionProofKey][]byte)
	 */
}

func (s *Server) GetPartitionForIdentifier(identifier []byte) *PartitionServer {
	hash := uint(binary.BigEndian.Uint64(libcrypto.Hash(identifier)))
	// if err != nil {
	// 	fmt.Printf("Error while getting partition for identifier: %v\n", err)
	// }
	// fmt.Printf("hash is %d\n", hash)
	return s.PartitionServers[hash%uint(len(s.PartitionServers))]
}

// Stores user key to a key-value store on the server.
func (s *Server) RegisterUserKey(ctx context.Context, user []byte,
	key []byte, signature []byte, verify bool) (uint64, error) {
	queryString := append(user, []byte("MK")...)
	partitionServer := s.GetPartitionForIdentifier(queryString)
	if verify {
		//verify the user is not registered already
		mk_serialized, _ := partitionServer.Storage.Get(ctx, queryString)
		if mk_serialized != nil {
			return 0, errors.New("User is already registered")
		}

		//verify self-signed masterkey
		if !crypto.VerifyBlob(key, signature, key) {
			return 0, errors.New("Verification failed")
		}
	}

	// Assign a position to the new entry
	partitionServer.LastPosLock.Lock()
	partitionServer.AppendLock.Lock()
	// fmt.Println("Register: last pos should be 0, it is ", partitionServer.LastPos)
	position := partitionServer.LastPos
	partitionServer.LastPos += 1
	// fmt.Println("Register: after incrementing, lastpos is ", partitionServer.LastPos)
	partitionServer.Partition.Append(user, queryString, key, signature)
	partitionServer.AppendLock.Unlock()
	partitionServer.LastPosLock.Unlock()

	// first username value pair, and thats the MK
	// in the estorage, it gets a separate key "usernameMK"

	serializedKey, _ := json.Marshal(
		ValueRecord{
			Position:  position,
			Signature: signature,
			Value:     key,
		})
	partitionServer.Storage.Put(ctx, queryString, serializedKey)
	return position, nil
}

/*

Akshit: we possibly don't need this function, as it seems to just be lookupmk and lookuppk but
for pk it takes in position so we need that part

// Retreives user key from a key-value store on the server.
// Does not return any proof associated with the key. This is a simple query op.
func (s *Server) GetUserKey(ctx context.Context, user []byte,
	masterKey bool, position uint64) ([]byte, []byte, uint64, error) {
	var key ValueRecord
	var keyList []ValueRecord
	queryString := user
	if masterKey {
		queryString = append(queryString, []byte("MK")...)
	} else {
		queryString = append(queryString, []byte("PK")...)
	}
	partitionServer := s.GetPartitionForIdentifier(queryString)

	serializedKey, err := partitionServer.Storage.Get(ctx, queryString)

	if err != nil {
		return nil, nil, 0, err
	}

	if serializedKey == nil {
		return nil, nil, 0, errors.New("no user key found")
	}

	partitionServer.epochLock.RLock()
	PublishedPos := partitionServer.PublishedPos
	partitionServer.epochLock.RUnlock()

	if masterKey {
		json.Unmarshal(serializedKey, &key)
	} else {
		json.Unmarshal(serializedKey, &keyList)
		// grab the latest key before provided position
		found := false
		for _, keyIter := range keyList {
			if keyIter.Position <= position && keyIter.Position <= PublishedPos {
				key = keyIter
				found = true
				break
			}
		}
		if !found || key.Position > position || key.Position > PublishedPos {
			// fmt.Println(key.Position)
			// fmt.Println(position)
			// fmt.Println(PublishedPos)
			return nil, nil, 0, errors.New("no key before requested position found")
		}
	}
	return key.Key, key.Signature, key.Position, nil
}
*/

// todo: modify constructor below for variable initialization
func NewServer(storage storage.Storage, updateEpochDuration time.Duration, numPartitions int, aggHistory bool, cfg *core.Config, tmpdir string) *Server {
	// privKey, err := vrf.GenerateKey(nil)
	// if err != nil {
	// 	return nil
	// }

	server := &Server{
		PartitionServers: []*PartitionServer{},

		epoch:               0,
		epochLock:           &sync.RWMutex{},
		updateEpochDuration: updateEpochDuration,
		verifyEpochDuration: cfg.VerificationPeriod,
		stopper:             make(chan struct{}),
	}

	for i := 0; i < numPartitions; i += 1 {
		var partitionServer *PartitionServer
		if !aggHistory {
			partitionServer = &PartitionServer{
				Partition:        core.NewPartition(),
				Storage:          storage,
				LastPos:          0,
				LastPosLock:      &sync.RWMutex{},
				NeedToRollUp:     false,
				NeedToRollUpLock: &sync.Mutex{},
				AppendLock:       &sync.Mutex{},
				Index:            i,
			}
		} else {
			partitionServer = &PartitionServer{
				Partition:        core.NewAggHistPartition(*cfg, tmpdir),
				Storage:          storage,
				LastPos:          0,
				LastPosLock:      &sync.RWMutex{},
				NeedToRollUp:     false,
				NeedToRollUpLock: &sync.Mutex{},
				AppendLock:       &sync.Mutex{},
				Index:            i,
			}
		}
		server.PartitionServers = append(server.PartitionServers, partitionServer)
	}

	// server.PublishedDigest = server.MerkleSquare.GetDigest()

	if updateEpochDuration != 0 {
		server.updateEpochDuration = updateEpochDuration
		go server.EpochLoop(time.Unix(0, time.Now().Add(updateEpochDuration).UnixNano()))
	}
	return server
}

func NewStoppedServer(storage storage.Storage, cfg *core.Config, tmpdir string) *Server {
	server := &Server{
		PartitionServers: []*PartitionServer{},

		epoch:               0,
		epochLock:           &sync.RWMutex{},
		updateEpochDuration: cfg.UpdatePeriod,
		verifyEpochDuration: cfg.VerificationPeriod,
		stopper:             make(chan struct{}),
	}

	for i := 0; i < int(cfg.Partitions); i += 1 {
		var partitionServer *PartitionServer
		if !cfg.AggHistory {
			partitionServer = &PartitionServer{
				Partition:        core.NewPartition(),
				Storage:          storage,
				LastPos:          0,
				LastPosLock:      &sync.RWMutex{},
				NeedToRollUp:     false,
				NeedToRollUpLock: &sync.Mutex{},
				AppendLock:       &sync.Mutex{},
				Index:            i,
			}
		} else {
			partitionServer = &PartitionServer{
				Partition:        core.NewAggHistPartition(*cfg, tmpdir),
				Storage:          storage,
				LastPos:          0,
				LastPosLock:      &sync.RWMutex{},
				NeedToRollUp:     false,
				NeedToRollUpLock: &sync.Mutex{},
				AppendLock:       &sync.Mutex{},
				Index:            i,
			}
		}
		server.PartitionServers = append(server.PartitionServers, partitionServer)
	}
	return server
}

func (s *Server) Start() {
	if s.updateEpochDuration != 0 {
		go s.EpochLoop(time.Unix(0, time.Now().Add(s.updateEpochDuration).UnixNano()))
	}
}

// NOT THREAD SAFE
func (s *Server) append(user []byte, id []byte, val []byte, sig []byte) {
	partitionServer := s.GetPartitionForIdentifier(id)

	position := partitionServer.LastPos
	partitionServer.LastPos += 1

	// Add to merkle tree
	partitionServer.Partition.Append(user, id, val, sig)

	// Add to KV store
	var serializedValue []byte
	// Prepend to existing entry
	ctx := context.Background()
	original, _ := partitionServer.Storage.Get(ctx, id)
	valueRecord := make([]ValueRecord, 1)
	valueRecord[0] = ValueRecord{
		Position:  position,
		Signature: sig,
		Value:     val,
	}
	if original == nil {
		serializedValue, _ = json.Marshal(valueRecord)
	} else {
		var deserialized []ValueRecord
		json.Unmarshal(original, &deserialized)
		serializedValue, _ = json.Marshal(append(valueRecord, deserialized...))
	}
	partitionServer.Storage.Put(ctx, id, serializedValue)
}

func (s *Server) PreloadServer(startIdx int, numAppends int, idSize int, valSize int) {
	masterSK := []byte{
		232, 197, 35, 104, 194, 130, 102, 207, 237, 150, 222, 125, 105, 185, 219, 217, 27, 243, 247, 40, 137, 252, 232, 107, 208, 104, 230, 160, 105, 179, 150, 61,
	}
	masterVK := []byte{
		62, 211, 17, 153, 196, 54, 94, 151, 176, 0, 248, 70, 42, 138, 114, 235, 1, 171, 213, 145, 8, 145, 100, 71, 109, 6, 10, 238, 69, 23, 4, 136,
	}
	//fmt.Println(masterSK)
	//fmt.Println(masterVK)
	sig := make([]byte, 64)
	value := generateRandomByteArray(valSize)
	crypto.SignBlob(masterSK, masterVK, sig, append(value, []byte("1")...))
	for i := 0; i < numAppends; i += 1 {
		id := []byte(strconv.Itoa(i + startIdx))
		s.append([]byte{}, id, value, sig)
		if i%1e5 == 0 {
			fmt.Println("Loaded", i, "entries")
		}
	}
}

func generateRandomByteArray(size int) []byte {
	res := make([]byte, size)

	_, err := rand.Read(res)
	if err != nil {
		panic(err)
	}

	return res
}

func (s *Server) SetupdateEpochDuration(newDuration time.Duration) {
	s.updateEpochDuration = newDuration
}

func (s *Server) EpochLoop(firstEpochCommitTime time.Time) {
	updateTimer := time.NewTimer(s.updateEpochDuration)
	defer updateTimer.Stop()
	var updateTicker *time.Ticker
	untilNextUpdateEpoch := updateTimer.C

	verificationTimer := time.NewTimer(s.verifyEpochDuration)
	defer verificationTimer.Stop()
	var verificationTicker *time.Ticker
	untilNextVerificationPeriod := verificationTimer.C

queryLoop:
	for {
		select {
		case <-untilNextUpdateEpoch:
			//NOTE: Debug
			if updateTicker == nil {
				fmt.Println("Starting update Ticker")
				updateTimer.Stop()
				updateTicker = time.NewTicker(s.updateEpochDuration)
				untilNextUpdateEpoch = updateTicker.C
			}
			log.Println("start incrementing update epoch")
			s.IncrementUpdateEpoch()
			log.Println("stop incrementing update epoch")

		case <-untilNextVerificationPeriod:
			if verificationTicker == nil {
				fmt.Println("Starting verification Ticker")
				verificationTimer.Stop()
				verificationTicker = time.NewTicker(s.verifyEpochDuration)
				untilNextVerificationPeriod = verificationTicker.C
			}
			log.Println("start incrementing verification epoch")
			s.IncrementVerificationPeriod()
			log.Println("stop incrementing verification epoch")

		case <-s.stopper:
			fmt.Println("Stopping epoch loop!")
			break queryLoop
		}
	}
	if updateTicker != nil {
		updateTicker.Stop()
	}
	if verificationTicker != nil {
		verificationTicker.Stop()
	}
}

// Stop ends the epoch loop. This is useful if you need to free all resources
// associated with a Server.
func (s *Server) Stop() {
	close(s.stopper)
}
