package bench

import (
	"context"
	"crypto/rand"

	"github.com/huyuncong/MerkleSquare/client"
	"github.com/huyuncong/MerkleSquare/constants"
	"github.com/huyuncong/MerkleSquare/verifier/verifierd"

	"github.com/immesys/bw2/crypto"

	coniks_crypto "github.com/coniks-sys/coniks-go/crypto"
)

const ServerAddr = "34.215.95.58" + constants.ServerPort

// const ServerAddr = "localhost" + constants.ServerPort
const AuditorAddr = "34.216.196.129" + constants.AuditorPort
const VerifierAddr = "localhost" + constants.VerifierPort

// type StorageClient interface {
// 	Register(ctx context.Context, username []byte, masterSK []byte, masterVK []byte) (uint64, error)
// 	Append(ctx context.Context, username []byte, key []byte) (uint64, error)
// 	LookUpMK(ctx context.Context, username []byte) ([]byte, uint64, error)
// 	LookUpPK(ctx context.Context, username []byte) ([]byte, uint64, error)
// }

//
// type StorageVerifier interface {
// 	// TODO
// }

// var TLSConfig = &tls.Config{InsecureSkipVerify: true}

func HelperNewClient(serverAddr string, auditorAddr string, verifierAddr string) (*client.Client, error) {
	return client.NewClient(serverAddr, auditorAddr, verifierAddr)
}

func HelperNewVerifier(serverAddr string) (*verifierd.Verifier, error) {
	return verifierd.NewBenchmarkVerifier(serverAddr)
}

// var seed uint32

func HelperMerkleSquareCreateKV(ctx context.Context, c *client.Client) ([]byte, error) {
	// myseed := atomic.AddUint32(&seed, 1)
	username := make([]byte, 32)
	if _, err := rand.Read(username); err != nil {
		return nil, err
	}
	masterSK, masterVK := crypto.GenerateKeypair()
	_, err := c.Register(ctx, username, masterSK, masterVK)
	return username, err
}

func HelperCreateKV(ctx context.Context, c *client.Client) (string, error) {
	var username []byte
	var err error

	username, err = HelperMerkleSquareCreateKV(ctx, c)

	return string(username), err
}

//==============================================================================

type Coniks_insert struct {
	Key   []byte
	Value []byte
	Index []byte
}

type Inserts struct {
	Key       []byte
	Value     []byte
	Signature []byte
}

var staticVRFKey = coniks_crypto.NewStaticTestVRFKey()

func GetKeyValueSigPairs(numPairs uint32) *[]Inserts {

	res := []Inserts{}

	for i := uint32(0); i < numPairs; i++ {

		key := generateRandomByteArray(keySize)
		value := generateRandomByteArray(valueSize)
		sig := generateRandomByteArray(sigSize)

		ins := Inserts{
			Key:       key,
			Value:     value,
			Signature: sig,
		}

		res = append(res, ins)
	}

	return &res
}

func GenerateRandomByteArray(size int) []byte {

	res := make([]byte, size)

	_, err := rand.Read(res)
	if err != nil {
		// handle error here
	}

	return res
}

func GetIndexKeyValuePairs(numPairs uint32) []Coniks_insert {

	res := []Coniks_insert{}

	for i := uint32(0); i < numPairs; i++ {

		key := GenerateRandomByteArray(keySize)
		index := staticVRFKey.Compute([]byte(key))
		value := GenerateRandomByteArray(valueSize)

		ins := Coniks_insert{
			Index: index,
			Key:   key,
			Value: value,
		}

		res = append(res, ins)
	}

	return res
}
