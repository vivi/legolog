package core

import (
	"fmt"
	"testing"
)

func TestHistoryForestAppend(t *testing.T) {
	forest := NewHistoryForest(31)
	tree0 := NewPrefixTree()
	tree1 := NewPrefixTree()
	tree1.PrefixAppend([]byte("1"), []byte("2"), 1)
	forest.Append(tree0)
	forest.Append(tree1)
	forest.Append(tree1)
	forest.Append(tree1)
	forest.Append(tree1)

	fmt.Println("tree0", tree0.root.getHash())
	fmt.Println("tree1", tree1.root.getHash())
	fmt.Println(forest.Roots[0].getHash())
	t.Error()
}
