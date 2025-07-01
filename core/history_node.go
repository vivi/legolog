package core

import (
	"encoding/binary"
	"fmt"

	crypto "github.com/huyuncong/MerkleSquare/lib/crypto"
)

// Base node struct
type historyNode struct {
	hash               []byte
	newestLeafHash     []byte
	verificationPeriod uint64
	parent             HistoryNode
	isRight            bool
	completed          bool
	index              index
}

// InternalHistoryNode in Merkle Tree
type InternalHistoryNode struct {
	historyNode
	leftChild  HistoryNode
	rightChild HistoryNode
}

// LeafHistoryNode representation in MerkleTree
type LeafHistoryNode struct {
	historyNode
}

// Called when a leaf node is created; fills in leaf node struct variables.
func (node *LeafHistoryNode) completeLeaf(verificationPeriod uint64, baseTreeHash []byte, pos uint32) {
	node.hash = baseTreeHash
	node.verificationPeriod = verificationPeriod
	node.newestLeafHash = node.hash
	node.completed = true
}

// Create ghost leaf node; call completeLeaf() to fill in other variables
func createLeafHistoryNode(parent HistoryNode, isRight bool, shift uint32) *LeafHistoryNode {
	return &LeafHistoryNode{
		historyNode: historyNode{
			parent:    parent,
			isRight:   isRight,
			completed: false,
			index: index{
				depth: 0,
				shift: shift,
			},
		},
	}
}

// Create ghost internal node; call internalNode.complete() to complete
func createInternalHistoryNode(parent HistoryNode, depth uint32, isRight bool, shift uint32) *InternalHistoryNode {
	return &InternalHistoryNode{
		historyNode: historyNode{
			parent:    parent,
			isRight:   isRight,
			completed: false,
			index: index{
				depth: depth,
				shift: shift,
			},
		},
	}
}

// Creates root node (called when initializing ghost nodes)
func createRootHistoryNode(depth uint32) HistoryNode {
	return &InternalHistoryNode{
		historyNode: historyNode{
			completed: false,
			index: index{
				depth: depth,
				shift: 0,
			},
		},
	}
}

//**********************************
// INTERFACE METHODS
//**********************************

type HistoryNode interface {
	isLeafNode() bool
	setParent(HistoryNode)
	getHash() []byte
	getNewestLeafHash() []byte
	complete()
	isComplete() bool
	isRightChild() bool
	getParent() HistoryNode
	createRightChild() HistoryNode
	createLeftChild() HistoryNode
	getRightChild() HistoryNode
	getLeftChild() HistoryNode
	getVerificationPeriod() uint64
	getDepth() uint32
	print()
	getShift() uint32
	getIndex() index
	getSibling() Sibling
	// serialize() ([]byte, error)
	getSize() int
}

func (node *InternalHistoryNode) complete() {
	hashVal := crypto.Hash(node.leftChild.getHash(), node.rightChild.getHash())

	node.hash = hashVal
	node.verificationPeriod = node.rightChild.getVerificationPeriod()
	node.newestLeafHash = node.rightChild.getNewestLeafHash()
	node.completed = true
}

// Called by internal nodes; creates a right child (of type internal or leaf)
func (node *InternalHistoryNode) createRightChild() HistoryNode {

	var newNode HistoryNode
	if node.getDepth() == 1 {
		newNode = createLeafHistoryNode(node, true, node.getShift()*2+1)
	} else {
		newNode = createInternalHistoryNode(node, node.getDepth()-1, true, node.getShift()*2+1)
	}

	node.rightChild = newNode

	return newNode
}

func (node *InternalHistoryNode) createLeftChild() HistoryNode {
	var newNode HistoryNode
	if node.getDepth() == 1 {
		newNode = createLeafHistoryNode(node, false, node.getShift()*2)
	} else {
		newNode = createInternalHistoryNode(node, node.getDepth()-1, false, node.getShift()*2)
	}

	node.leftChild = newNode

	return newNode
}

func (node *InternalHistoryNode) getSibling() Sibling {

	var hash []byte
	if node.isRightChild() {
		hash = node.getParent().getLeftChild().getHash()
	} else {
		hash = node.getParent().getRightChild().getHash()
	}

	return Sibling{
		Hash: hash,
	}
}

func (node *LeafHistoryNode) getSibling() Sibling {

	var hash []byte
	if node.isRightChild() {
		hash = node.getParent().getLeftChild().getHash()
	} else {
		hash = node.getParent().getRightChild().getHash()
	}

	return Sibling{
		Hash: hash,
	}
}

func (node *LeafHistoryNode) getSize() int {

	// pointer to parent 8 bytes
	total := pointerSizeInBytes

	// hash, isRight, isComplete sizes
	total += binary.Size(node.hash) + binary.Size(node.newestLeafHash) + binary.Size(node.isRight) + binary.Size(node.isComplete)

	// size of index
	total += binary.Size(node.index.depth) + binary.Size(node.index.shift)

	return total
}

func (node *InternalHistoryNode) getSize() int {
	// pointer to parent 8 bytes + left and right child pointers + prefix tree pointer
	total := pointerSizeInBytes * 4

	// hash, isRight, isComplete sizes
	total += binary.Size(node.hash) + binary.Size(node.newestLeafHash) + binary.Size(node.isRight) + binary.Size(node.isComplete)

	// size of index
	total += binary.Size(node.index.depth) + binary.Size(node.index.shift)

	// right child
	if node.getRightChild() != nil {
		total += node.getRightChild().getSize()
	}

	if node.getLeftChild() != nil {
		total += node.getLeftChild().getSize()
	}

	// fmt.Println(node.getPrefixTree().getSize())
	// fmt.Println(pointerSizeInBytes)
	// fmt.Println(binary.Size(node.hash))

	return total
}

func (node *InternalHistoryNode) isComplete() bool              { return node.completed }
func (node *InternalHistoryNode) isRightChild() bool            { return node.isRight }
func (node *InternalHistoryNode) getParent() HistoryNode        { return node.parent }
func (node *InternalHistoryNode) isLeafNode() bool              { return false }
func (node *InternalHistoryNode) setParent(parent HistoryNode)  { node.parent = parent }
func (node *InternalHistoryNode) getHash() []byte               { return node.hash }
func (node *InternalHistoryNode) getNewestLeafHash() []byte     { return node.newestLeafHash }
func (node *InternalHistoryNode) getVerificationPeriod() uint64 { return node.verificationPeriod }
func (node *InternalHistoryNode) getRightChild() HistoryNode    { return node.rightChild }
func (node *InternalHistoryNode) getLeftChild() HistoryNode     { return node.leftChild }
func (node *InternalHistoryNode) getDepth() uint32              { return node.index.depth }
func (node *InternalHistoryNode) print()                        { fmt.Print(node.isComplete()) }
func (node *InternalHistoryNode) getShift() uint32              { return node.index.shift }
func (node *InternalHistoryNode) getIndex() index               { return node.index }

func (node *LeafHistoryNode) isComplete() bool              { return node.completed }
func (node *LeafHistoryNode) isRightChild() bool            { return node.isRight }
func (node *LeafHistoryNode) getParent() HistoryNode        { return node.parent }
func (node *LeafHistoryNode) isLeafNode() bool              { return true }
func (node *LeafHistoryNode) setParent(parent HistoryNode)  { node.parent = parent }
func (node *LeafHistoryNode) getHash() []byte               { return node.hash }
func (node *LeafHistoryNode) getNewestLeafHash() []byte     { return node.newestLeafHash }
func (node *LeafHistoryNode) getVerificationPeriod() uint64 { return node.verificationPeriod }
func (node *LeafHistoryNode) complete()                     {}
func (node *LeafHistoryNode) createLeftChild() HistoryNode  { return &LeafHistoryNode{} }
func (node *LeafHistoryNode) createRightChild() HistoryNode { return &LeafHistoryNode{} }
func (node *LeafHistoryNode) getRightChild() HistoryNode    { return &LeafHistoryNode{} }
func (node *LeafHistoryNode) getLeftChild() HistoryNode     { return &LeafHistoryNode{} }
func (node *LeafHistoryNode) getDepth() uint32              { return 0 }
func (node *LeafHistoryNode) print()                        { fmt.Print(node.isComplete()) }
func (node *LeafHistoryNode) getShift() uint32              { return node.index.shift }
func (node *LeafHistoryNode) getIndex() index               { return node.index }
