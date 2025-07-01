package core

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"

	crypto "github.com/huyuncong/MerkleSquare/lib/crypto"
)

type prefixAppend struct {
	Prefix []byte `json:"prefix"`
	Value  []byte `json:"value"`
	Pos    uint32 `json:"pos"`
}

// TODO: Maybe we need to make it public in the future. We can decide later
// -- Yuncong
type prefixTree struct {
	root       *internalNode
	isComplete bool
	appends    []prefixAppend
}

// for node on path to root, store onpath partial prefix, and the hash of the offpath child,
type forNodeOnCopath struct {
	// for root there is no partial prefix
	PartialPrefix []byte
	//the child node that isn't on path (struct for starting node stores node itself)
	OtherChildHash []byte
}

// MembershipProof ...
type MembershipProof struct {
	LeafPartialPrefix []byte
	CopathNodes       []forNodeOnCopath //first is leaf's sibling, last is root
}

// NonMembershipProof ...
type NonMembershipProof struct {
	EndNodeHash          []byte            // for empty nodes, will be nil
	EndNodePartialPrefix []byte            // for empty nodes, will be the (one) next-expected byte
	CopathNodes          []forNodeOnCopath //first is node at bottom of path, last is root
}

func NewPrefixTree() *prefixTree {

	res := &prefixTree{
		root: &internalNode{
			parent:        nil,
			hash:          nil,
			leftChild:     nil,
			rightChild:    nil,
			partialPrefix: nil,
		},
		isComplete: false,
	}

	return res
}

func makePrefixFromKey(key []byte) []byte {
	return ConvertBitsToBytes(crypto.Hash(key))
}
func MakePrefixFromKey(key []byte) []byte {
	return ConvertBitsToBytes(crypto.Hash(key))
}

func ConvertBitsToBytes(asBits []byte) []byte {

	res := []byte{}

	for _, byt := range asBits {
		for i := 0; i < 8; i++ {
			if ((1 << (7 - i)) & byt) == 0 {
				res = append(res, 0)
			} else {
				res = append(res, 1)
			}
		}
	}

	return res
}

func getBit(prefix []byte, i uint32) byte {
	return (prefix[i/8] >> (7 - i%8)) & 1
}

func (tree *prefixTree) PrefixAppend(prefix []byte, valueHash []byte, pos uint32) (err error) { //// TODO: add hashes and parents

	if tree.isComplete {
		err = errors.New("cannot append to completed prefix tree")
		return
	}

	tree.appends = append(tree.appends, prefixAppend{prefix, valueHash, pos})

	var prev prefixNode
	var curr prefixNode = tree.root
	i := uint32(0)

	for i < uint32(len(prefix)) {

		prev = curr
		curr = curr.getChild(prefix, i)

		if curr == nil { //this child had not been made yet
			leaf := newLeafNode(prev, valueHash, pos, prefix[i:])
			tree.updateHashesFromLeaf(leaf)
			return
		}

		j := uint32(0)
		for j < uint32(len(curr.getPartialPrefix())) {
			if prefix[i] == curr.getPartialPrefix()[j] {
				i++
				j++
			} else {
				newParent := splitCompressedNode(curr, prev, j)
				curr.updateHash()
				leaf := newLeafNode(newParent, valueHash, pos, prefix[i:])
				tree.updateHashesFromLeaf(leaf)
				return
			}
		}
	}
	leaf := curr
	leaf.addValue(valueHash, pos)
	tree.updateHashesFromLeaf(leaf)
	return
}

func (tree *prefixTree) getLeaf(prefix []byte) prefixNode {

	var curr prefixNode = tree.root
	i := uint32(0)

	for i < uint32(len(prefix)) {

		curr = curr.getChild(prefix, i)
		if curr == nil {
			return nil //key doesn't exist in tree
		}
		partialPrefix := curr.getPartialPrefix()
		if bytes.Compare(prefix[i:i+uint32(len(partialPrefix))], partialPrefix) == 0 {
			i += uint32(len(partialPrefix))
			continue
		} else {
			return nil //key doesn't exist in tree
		}
	}
	return curr

}

func (tree *prefixTree) generateMembershipProof(prefix []byte) (proof *MembershipProof, leafValues []KeyHash) {
	leaf := tree.getLeaf(prefix)
	if leaf == nil {
		return nil, nil
	}

	if !leaf.isLeafNode() {
		panic("prefix path should end with a leaf, but does not")
	}
	copath := tree.buildCopathFromNode(leaf)
	return &MembershipProof{
		LeafPartialPrefix: leaf.getPartialPrefix(),
		CopathNodes:       copath,
	}, leaf.getValues()
}

func (tree *prefixTree) generateNonMembershipProof(prefix []byte) *NonMembershipProof {

	var prev prefixNode
	var curr prefixNode = tree.root
	i := uint32(0)

	for i < uint32(len(prefix)) {

		prev = curr
		curr = curr.getChild(prefix, i)

		if curr == nil {
			if prev != tree.root {
				panic("root should be the only internal node that can have <2 children")
			}
			missingNode := &internalNode{
				parent:        tree.root,
				partialPrefix: []byte{prefix[i]},
			}
			return &NonMembershipProof{
				EndNodeHash:          nil,
				EndNodePartialPrefix: missingNode.getPartialPrefix(),
				CopathNodes:          tree.buildCopathFromNode(missingNode),
			}
		}
		partialPrefix := curr.getPartialPrefix()
		if bytes.Compare(prefix[i:i+uint32(len(partialPrefix))], partialPrefix) == 0 {
			i += uint32(len(partialPrefix))
			continue
		} else {
			// WARNING: Again, this seems to miss something.
			// At least, you should use "i" or "j" information and recalculate the hash
			// as if there is a splited node at "i" or "j".
			// TODO: discuss with Yuncong again, I don't believe faking a split is necessary
			return &NonMembershipProof{
				EndNodeHash:          curr.getHash(),
				EndNodePartialPrefix: curr.getPartialPrefix(),
				CopathNodes:          tree.buildCopathFromNode(curr),
			}
		}
	}
	return nil //key exists
}

func (tree *prefixTree) buildCopathFromNode(startingNode prefixNode) []forNodeOnCopath {
	copath := []forNodeOnCopath{}
	curr := startingNode
	for curr.getParent() != nil {
		var siblingHash []byte
		if curr.getSibling() != nil {
			siblingHash = curr.getSibling().getHash()
		}
		copath = append(copath,
			forNodeOnCopath{
				PartialPrefix:  curr.getParent().getPartialPrefix(),
				OtherChildHash: siblingHash,
			})
		curr = curr.getParent()
	}
	if curr != tree.root {
		panic("copath should end at root, there is a node on path missing a parent value")
	}

	return copath
}

func splitCompressedNode(nodeToSplit prefixNode, parent prefixNode, index uint32) *internalNode {
	prefixLength := uint32(len(nodeToSplit.getPartialPrefix()))
	if prefixLength <= 1 {
		panic("can't split a non-compressed node")
	} else if index == 0 || index >= prefixLength {
		panic("given index doesn't split the prefix into 2 peices")
	}
	intermediateNode := newInteriorNode(parent, nodeToSplit.getPartialPrefix()[0:index])

	nodeToSplit.setPartialPrefix(nodeToSplit.getPartialPrefix()[index:])
	intermediateNode.addChild(nodeToSplit)

	return intermediateNode
}

func (tree *prefixTree) updateHashesFromLeaf(leaf prefixNode) {

	if !leaf.isLeafNode() {
		panic("updateHashesFromLeaf was passed internalNode as argument")
	}

	curr := leaf
	for curr != tree.root {
		curr.updateHash()
		curr = curr.getParent()
	}
	tree.root.updateHash()
}

func (tree *prefixTree) getHash() []byte {
	return tree.root.hash
}

func (tree *prefixTree) complete() {
	tree.isComplete = true
}

type merkleProof struct {
	endNodeHash          []byte
	endNodePartialPrefix []byte
	copath               []forNodeOnCopath
}

func getPrefix(copath []forNodeOnCopath) []byte {
	prefixInProof := []byte{}
	for i := len(copath) - 1; i >= 0; i-- {
		prefixInProof = append(prefixInProof, copath[i].PartialPrefix...)
	}
	return prefixInProof
}

// calculating hashes along the copath gives same root hash as expected
func getRootHash(endNodeHash []byte, endNodePartialPrefix []byte, copath []forNodeOnCopath) []byte {
	currHash := endNodeHash
	comingFromLeft := endNodePartialPrefix[0] == 0
	var leftHash, rightHash []byte
	for i, nodeOnCopath := range copath {
		if i != len(copath)-1 { // not root
			if nodeOnCopath.OtherChildHash == nil {
				panic("there is an empty node in copath, that isn't a child of root")
			} else if nodeOnCopath.PartialPrefix == nil {
				panic("all nodes other than root should have a partial prefix")
			}
		}
		if comingFromLeft {
			leftHash = currHash
			rightHash = nodeOnCopath.OtherChildHash
		} else {
			leftHash = nodeOnCopath.OtherChildHash
			rightHash = currHash
		}
		//nodeOnCopath in the while loop will always be an internal node
		currHash = crypto.Hash(nodeOnCopath.PartialPrefix, leftHash, rightHash)
		if i != len(copath)-1 { //otherwise is is root
			comingFromLeft = nodeOnCopath.PartialPrefix[0] == 0
		}
	}
	return currHash
}

// NOTE: For the verification, you should provide key-value information.
// For verifyMembershipProof, the input should contain all the key-value pairs in the leaf node,
// including positions and signatures.
// TODO: will add this after key-value store implemented
func computeRootHashMembership(prefix []byte, proof *MembershipProof, leafValues []KeyHash) (rootHash []byte) {
	if !bytes.Equal(prefix, append(getPrefix(proof.CopathNodes), proof.LeafPartialPrefix...)) {
		return nil //copath in proof leads somewhere other than key's leaf node
	}
	return getRootHash(leafHash(proof.LeafPartialPrefix, leafValues), proof.LeafPartialPrefix, proof.CopathNodes)
}

func computeRootHashNonMembership(prefix []byte, proof *NonMembershipProof) (rootHash []byte) {
	copathPartialPrefix := getPrefix(proof.CopathNodes)
	if !bytes.HasPrefix(prefix, copathPartialPrefix) {
		return nil //copath forms prefix that isn't a frontal partial slice of prefix=crypto.Hash(key)
	}
	remainingPrefix := prefix[len(copathPartialPrefix):]
	if remainingPrefix[0] != proof.EndNodePartialPrefix[0] {
		return nil //proof's endNode is in the copath (sibling to on-path node) instead of on-path to key
	}
	if proof.EndNodeHash == nil { //should be an empty node under root
		if len(proof.CopathNodes) != 1 {
			panic("endNode is an empty node that isn't a child of root (endNodeHash should only be nil for empty nodes, empty node can only exist as children of root, and copath should have len==1 if endNode is child of root)")
		}
	} else { // a compressed node exists that would've been split if prefix was in the tree
		if bytes.HasPrefix(remainingPrefix, proof.EndNodePartialPrefix) {
			return nil //proof's endNode is a compressed node whose partial prefix matches key's path, key could exist under endNode
		}
	}
	return getRootHash(proof.EndNodeHash, proof.EndNodePartialPrefix, proof.CopathNodes)
}

func (tree *prefixTree) getSize() int {

	// pointer to root
	total := pointerSizeInBytes

	// isComplete bool
	total += binary.Size(tree.isComplete)
	// fmt.Println(binary.Size(tree.isComplete))

	// actual tree size
	total += tree.root.getSize()

	// sum := tree.root.getNumNodes()
	// fmt.Println(sum)
	// fmt.Println(total)

	return total
}

func (tree *prefixTree) copy() (*prefixTree, error) {
	ret := NewPrefixTree()
	for _, append := range tree.appends {
		err := ret.PrefixAppend(append.Prefix, append.Value, append.Pos)
		if err != nil {
			return nil, err
		}
	}
	return ret, nil

	treeJson, err := tree.serialize()
	if err != nil {
		return nil, err
	}
	ret, err = deserializePrefixTree(treeJson)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func (p *prefixTree) HasValue(prefix []byte, value []byte) bool {
	if p.getLeaf(prefix) == nil {
		return false
	}
	prefixNode := p.getLeaf(prefix)
	for _, entry := range prefixNode.getValues() {
		if bytes.Equal(value, entry.Hash) {
			return true
		}
	}
	return false
}
func (p *prefixTree) String() string {
	currLevel := []prefixNode{p.root}
	ret := ""
	level := 0
	for len(currLevel) > 0 {
		ret += fmt.Sprintf("Level %d: ", level)
		nextLevel := []prefixNode{}
		for _, node := range currLevel {
			if node.getParent() != nil {
				ret += fmt.Sprintf("%d<-", node.getParent().getPartialPrefix())
			}
			ret += fmt.Sprintf("%x\t", node.getHash()[:10])
			if node.getLeftChild()!= nil {
				nextLevel = append(nextLevel, node.getLeftChild())
			}
			if node.getRightChild() != nil {
				nextLevel = append(nextLevel, node.getRightChild())
			}
		}
		ret += "\n"
		currLevel = nextLevel
		level += 1
	}
	return ret
}