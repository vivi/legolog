package bench

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math/rand"
	"sort"
	"testing"

	"github.com/huyuncong/MerkleSquare/core"
)

type insert struct {
	Key       []byte
	Value     []byte
	Signature []byte
}

const keySize = 32
const valueSize = 32
const sigSize = 32

//*******************************
// SETUP METHODS
//*******************************

// createAndStoreTree should be run before all benchmarks to serialize Merkle Square objects
// to reduce building time before proof benchmarks
func createAndStoreTree(n int) error {

	values := getKeyValueSigPairs(1 << n)
	ms := createTestingTreeRandomValues(20, values)

	buf, err := ms.Serialize()
	if err != nil {
		return err
	}

	treeBytes, err := core.WriteBytesToFile(buf, fmt.Sprintf("/fixedtrees/FixedDepth%d", 1<<n))
	fmt.Printf("Total tree bytes %d", treeBytes)
	if err != nil {
		return err
	}

	buf, err = serializeInserts(values)
	if err != nil {
		return err
	}

	insertBytes, err := core.WriteBytesToFile(buf, fmt.Sprintf("/fixedinserts/FixedDepthInserts%d", 1<<n))
	fmt.Printf("Total insert bytes %d", insertBytes)
	if err != nil {
		return err
	}

	return nil
}

// loadTreeAndKeys loads the tree corresponding to 2^n size and returns the
// key value pairs it is comprised of
func loadTreeAndKeys(n int) (*core.MerkleSquare, *[]insert, error) {

	insertsBuf, err := core.ReadBytesFromFile(fmt.Sprintf("/fixedinserts/FixedDepthInserts%d", 1<<n))
	if err != nil {
		return nil, nil, err
	}

	inserts, err := deserializeInserts(insertsBuf)
	if err != nil {
		return nil, nil, err
	}

	msBuf, err := core.ReadBytesFromFile(fmt.Sprintf("/fixedtrees/FixedDepth%d", 1<<n))
	if err != nil {
		return nil, nil, err
	}

	ms, err := core.DeserializeMerkleSquare(msBuf)
	if err != nil {
		return nil, nil, err
	}

	return ms, inserts, nil
}

//*******************************
// HELPER METHODS
//*******************************

func createRepeatedValues(inserts *[]insert, numRepeats int) ([]byte, []uint32) {

	lenValues := len(*inserts)
	values := map[int32]int{}
	count := 1

	randPos := rand.Int31n(int32(lenValues))
	values[randPos] = -1
	key := (*inserts)[randPos].Key
	keyPositions := []int{int(randPos)}

	for count < numRepeats {
		val := rand.Int31n(int32(lenValues))

		if values[val] != -1 {

			(*inserts)[val].Key = key

			values[val] = -1
			keyPositions = append(keyPositions, int(val))
			count++
		}
	}

	sort.Ints(keyPositions)

	res := []uint32{}

	for _, num := range keyPositions {
		res = append(res, uint32(num))
	}

	return key, res
}

func getKeyHashesFromInserts(inserts *[]insert, keyPositions *[]uint32) *[]core.KeyHash {

	res := []core.KeyHash{}

	for _, pos := range *keyPositions {
		ins := (*inserts)[pos]
		keyHash := core.KeyHash{
			Hash: core.ComputeContentHash(ins.Key, ins.Value, ins.Signature, pos),
			Pos:  pos,
		}

		res = append(res, keyHash)
	}

	return &res
}

func checkNonexistenceProofSize(proof *core.NonExistenceProof) int {

	total := 0

	for _, nonMemProof := range proof.NonMembershipProofs {
		total += getNonMemProofSize(&nonMemProof)
	}

	for _, childHash := range proof.ChildHashes {
		total += binary.Size(childHash)
	}

	total += checkExtensionProofSize(&proof.ExtensionProof)

	total += getLeafHashSize(&proof.LeafHash)

	return total
}

func checkLatestPKProofSize(proof *core.LatestPKProof) int {

	total := 0

	for _, nonMemProof := range proof.NonMembershipProofs {
		total += getNonMemProofSize(&nonMemProof)
	}

	total += getMemProofSize(&proof.MembershipProof)

	for _, childHash := range proof.ChildHashes {
		total += binary.Size(childHash)
	}

	total += getLeafHashSize(&proof.LeafHash)

	for _, keyHash := range proof.OtherHashes {
		total += getKeyHashSize(&keyHash)
	}

	return total
}

func checkMKProofSize(proof *core.MKProof) int {

	total := 0

	for _, nonMemProof := range proof.NonMembershipProofs {
		total += getNonMemProofSize(&nonMemProof)
	}

	total += getMemProofSize(&proof.MembershipProof)

	for _, childHash := range proof.ChildHashes {
		total += binary.Size(childHash)
	}

	total += getLeafHashSize(&proof.LeafHash)

	for _, keyHash := range proof.OtherHashes {
		total += getKeyHashSize(&keyHash)
	}

	return total
}

func checkExtensionProofSize(proof *core.MerkleExtensionProof) int {

	total := 0

	for _, sib := range proof.Siblings {
		total += getSiblingSize(&sib)
	}

	for _, prefixHash := range proof.PrefixHashes {
		total += binary.Size(prefixHash)
	}

	return total
}

func checkExistenceProofSize(proof *core.MerkleExistenceProof) int {

	total := 0

	for _, sib := range proof.Siblings {
		total += getSiblingSize(&sib)
	}

	for _, memProof := range proof.PrefixProofs {
		total += getMemProofSize(&memProof)
	}

	return total
}

func checkBatchedLookupProofSize(proof *core.BatchedLookupProof) int {

	total := 0

	for _, memProof := range proof.MembershipProofs {
		total += getMemProofSize(&memProof)
	}

	for _, nonMemProof := range proof.NonMemberShipProofs {
		total += getNonMemProofSize(&nonMemProof)
	}

	for _, keyHash := range proof.OtherHashes {
		total += getKeyHashSize(&keyHash)
	}

	for _, childHash := range proof.ChildHashes {
		total += binary.Size(childHash)
	}

	total += getLeafHashSize(&proof.LeafHash)
	total += binary.Size(proof.MemProofIndeces)

	return total
}

func getSiblingSize(sib *core.Sibling) int {
	return binary.Size(sib.Hash)
}

func getMemProofSize(memProof *core.MembershipProof) int {
	total := 0

	total += binary.Size(memProof.LeafPartialPrefix) / 8

	for _, copathNode := range memProof.CopathNodes {
		total += binary.Size(copathNode.OtherChildHash) + binary.Size(copathNode.PartialPrefix)/8
	}

	return total
}

func getNonMemProofSize(nonMemProof *core.NonMembershipProof) int {

	total := 0

	total += binary.Size(nonMemProof.EndNodeHash) + binary.Size(nonMemProof.EndNodePartialPrefix)/8

	for _, copathNode := range nonMemProof.CopathNodes {
		total += binary.Size(copathNode.OtherChildHash) + binary.Size(copathNode.PartialPrefix)/8
	}

	return total
}

func getLeafHashSize(leafHash *core.LeafHash) int {
	//return binary.Size(leafHash.Prefix) + binary.Size(leafHash.NodeContentHash)
	return binary.Size(leafHash.Prefix)/8 + binary.Size(leafHash.NodeContentHash)
}

func getKeyHashSize(keyHash *core.KeyHash) int {
	return binary.Size(keyHash.Hash) + binary.Size(keyHash.Pos)
}

func createTestingTreeRandomValues(depth uint32, inserts *[]insert) *core.MerkleSquare {

	m := core.NewMerkleSquare(depth)

	for _, ins := range *inserts {
		m.Append(ins.Key, ins.Value, ins.Signature)
	}

	return m
}

func getKeyValueSigPairs(numPairs uint32) *[]insert {

	res := []insert{}

	for i := uint32(0); i < numPairs; i++ {

		key := generateRandomByteArray(keySize)
		value := generateRandomByteArray(valueSize)
		sig := generateRandomByteArray(sigSize)

		ins := insert{
			Key:       key,
			Value:     value,
			Signature: sig,
		}

		res = append(res, ins)
	}

	return &res
}

func generateRandomByteArray(size int) []byte {

	res := make([]byte, size)

	_, err := rand.Read(res)
	if err != nil {
		panic(err)
	}

	return res
}

func writeBenchmarkResultToFile(res *testing.BenchmarkResult, fileName string) error {

	buf, err := json.Marshal(*res)
	if err != nil {
		return err
	}

	_, err = core.WriteBytesToFile(buf, fileName)

	return err
}

//*******************************
// INSERT SERIALIZATION
//*******************************

func serializeInserts(inserts *[]insert) ([]byte, error) {

	buf, err := json.Marshal(inserts)
	if err != nil {
		return nil, err
	}

	return buf, nil
}

func deserializeInserts(buf []byte) (*[]insert, error) {

	var inserts []insert
	err := json.Unmarshal(buf, &inserts)

	if err != nil {
		return nil, err
	}

	return &inserts, nil
}
