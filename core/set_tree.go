package core

import (
	"bytes"
	"errors"
	"fmt"
)

/*
* A set tree is a prefix tree that can prove inclusion and exclusion.
* See https://people.csail.mit.edu/silvio/Selected%20Scientific%20Papers/Zero%20Knowledge/Zero-Knowledge_Sets.pdf
 */

type SetTree struct {
	tree       *prefixTree
	prefix_len int
}

func NewSetTree(prefixes [][]byte, valHashes [][]byte) (res *SetTree, err error) {
	if len(prefixes) != len(valHashes) {
		err = errors.New("prefixes and valHashes are of unequal length")
	}
	//prefixLen := int(math.Ceil(math.Log2(float64(len(prefixes)))))
	prefixLen := len(prefixes[0])

	// prefixes = [H(identifier1), H(identifier2), ...]
	res = &SetTree{
		tree:       NewPrefixTree(),
		prefix_len: prefixLen,
	}

	for i := range prefixes {
		err = res.tree.PrefixAppend(prefixes[i][:prefixLen], valHashes[i], 1)
		if err != nil {
			return
		}
	}
	frontier := computeFrontier(res.tree.root, []byte{}, 1, prefixLen) // TODO: Check depth is not off-by-one here
	for _, item := range frontier {
		err = res.tree.PrefixAppend(item, item, 0)
		if err != nil {
			return
		}
	}
	return
}

func computeFrontier(curr prefixNode, prefixSoFar []byte, depth int, maxDepth int) (res [][]byte) {
	if depth == maxDepth {
		return
	}
	partialPrefix := curr.getPartialPrefix()
	if len(partialPrefix) > 1 {
		// We are in the compressed part of the prefix tree
		for _, bit := range partialPrefix {
			// The other bit is a part of the frontier
			frontierBit := 1 - bit
			frontierHash := append(prefixSoFar, frontierBit)
			res = append(res, frontierHash)

			prefixSoFar = append(prefixSoFar, bit)
		}
		depth += len(partialPrefix) - 1
	} else if len(partialPrefix) == 1 {
		prefixSoFar = append(prefixSoFar, partialPrefix[0])
	}
	if curr.getLeftChild() == nil {
		frontierHash := append(prefixSoFar, 0)
		res = append(res, frontierHash)
	} else {
		computeFrontier(curr.getLeftChild(), prefixSoFar, depth+1, maxDepth)
	}
	if curr.getRightChild() == nil {
		frontierHash := append(prefixSoFar, 1)
		res = append(res, frontierHash)
	} else {
		computeFrontier(curr.getRightChild(), prefixSoFar, depth+1, maxDepth)
	}
	return
}

func (s *SetTree) ProveExistence(prefix []byte) (proof *MembershipProof, leafValues []KeyHash) {
	prefix = prefix[:s.prefix_len]
	return s.tree.generateMembershipProof(prefix)
}

func (s *SetTree) ProveNonExistence(prefix []byte) (proof *MembershipProof, leafValues []KeyHash) {
	/*
		XXX: This probably doesn't need any special logic for handling compression because computeFrontier should expand
			any compressed nodes that are not in the set.
	*/
	var curr prefixNode = s.tree.root
	var prefixToProve []byte
	for _, bit := range prefix {
		if curr.getRightChild() == nil && curr.getLeftChild() == nil {
			break
		}
		if bit == 0 {
			curr = curr.getLeftChild()
		} else {
			curr = curr.getRightChild()
		}
		prefixToProve = append(prefixToProve, bit)
	}
	fmt.Println("prefixToProve", prefixToProve)
	return s.tree.generateMembershipProof(prefixToProve)
}

func (s *SetTree) HasKey(prefix []byte) bool {
	prefix = prefix[:s.prefix_len]
	return s.tree.getLeaf(prefix) != nil
}

func (s *SetTree) HasValue(prefix []byte, value []byte) bool {
	if !s.HasKey(prefix) {
		return false
	}
	prefixNode := s.GetLeaf(prefix)
	for _, entry := range prefixNode.getValues() {
		if bytes.Equal(value, entry.Hash) {
			return true
		}
	}
	return false

}

func (s *SetTree) GetLeaf(prefix []byte) prefixNode {
	prefix = prefix[:s.prefix_len]
	return s.tree.getLeaf(prefix)
}
