package core

import (
	"bytes"
	"testing"
)

func TestSimpleSetTree(t *testing.T) {
	prefixes := [][]byte{{0b0, 0b0}, {0b0, 0b1}, {0b1, 0b0}, {0b1, 0b1}}
	valHashes := [][]byte{{0b0, 0b0}, {0b0, 0b1}, {0b1, 0b0}, {0b1, 0b1}}
	setTree, _ := NewSetTree(prefixes, valHashes)

	CheckSetTreeConstruction(t, setTree, prefixes, valHashes)
}

func TestConflict(t *testing.T) {
	prefixes := [][]byte{{0b0}, {0b0}}
	valHashes := [][]byte{{0b0, 0b0}, {0b0, 0b1}}
	setTree, _ := NewSetTree(prefixes, valHashes)

	CheckSetTreeConstruction(t, setTree, prefixes, valHashes)
}

func TestCompressed(t *testing.T) {
	prefixes := [][]byte{{0b0, 0b0, 0b1}, {0b0, 0b0, 0b0}}
	valHashes := [][]byte{{0b0, 0b0}, {0b0, 0b1}}
	setTree, _ := NewSetTree(prefixes, valHashes)

	CheckSetTreeConstruction(t, setTree, prefixes, valHashes)
}
func TestLongerPrefix(t *testing.T) {
	prefixes := [][]byte{{0b0, 0b1}, {0b1, 0b0}}
	valHashes := [][]byte{{0b0, 0b0}, {0b0, 0b1}}
	setTree, _ := NewSetTree(prefixes, valHashes)

	CheckSetTreeConstruction(t, setTree, prefixes, valHashes)
}

func TestFullHashLengthPrefixes(t *testing.T) {
	prefixes := [][]byte{{0b0, 0b1, 0b0, 0b0, 0b0, 0b0, 0b1, 0b1, 0b1, 0b0, 0b0, 0b0, 0b0, 0b1, 0b1, 0b0, 0b1, 0b0, 0b1, 0b0, 0b1, 0b1, 0b1, 0b0, 0b0, 0b0, 0b1, 0b0, 0b1, 0b1, 0b1, 0b0, 0b1, 0b0, 0b0, 0b1, 0b1, 0b0, 0b0, 0b1, 0b1, 0b0, 0b0, 0b1, 0b1, 0b0, 0b1, 0b0, 0b1, 0b1, 0b0, 0b0, 0b0, 0b0, 0b1, 0b1, 0b1, 0b0, 0b1, 0b1, 0b1, 0b0, 0b1, 0b1, 0b1, 0b0, 0b1, 0b1, 0b1, 0b1, 0b0, 0b1, 0b1, 0b1, 0b1, 0b1, 0b0, 0b0, 0b0, 0b1, 0b0, 0b1, 0b0, 0b0, 0b1, 0b1, 0b1, 0b0, 0b1, 0b1, 0b0, 0b1, 0b1, 0b0, 0b1, 0b0, 0b0, 0b0, 0b1, 0b0, 0b0, 0b0, 0b1, 0b0, 0b1, 0b1, 0b0, 0b1, 0b1, 0b1, 0b1, 0b0, 0b0, 0b1, 0b0, 0b0, 0b1, 0b1, 0b1, 0b1, 0b1, 0b0, 0b1, 0b1, 0b0, 0b0, 0b0, 0b1, 0b0, 0b0, 0b1, 0b1, 0b0, 0b1, 0b1, 0b1, 0b1, 0b0, 0b0, 0b1, 0b0, 0b0, 0b0, 0b0, 0b0, 0b1, 0b1, 0b0, 0b0, 0b0, 0b0, 0b0, 0b1, 0b1, 0b0, 0b0, 0b1, 0b0, 0b1, 0b0, 0b0, 0b1, 0b1, 0b1, 0b0, 0b1, 0b0, 0b0, 0b1, 0b0, 0b0, 0b0, 0b1, 0b0, 0b1, 0b1, 0b0, 0b1, 0b0, 0b0, 0b1, 0b0, 0b1, 0b1, 0b1, 0b1, 0b0, 0b0, 0b1, 0b1, 0b0, 0b0, 0b1, 0b1, 0b0, 0b0, 0b0, 0b1, 0b1, 0b0, 0b1, 0b1, 0b1, 0b0, 0b1, 0b1, 0b1, 0b0, 0b0, 0b1, 0b1, 0b0, 0b0, 0b1, 0b0, 0b1, 0b0, 0b0, 0b1, 0b0, 0b0, 0b1, 0b1, 0b0, 0b0, 0b0, 0b1, 0b0, 0b0, 0b1, 0b1, 0b0, 0b1, 0b0, 0b0, 0b1, 0b1, 0b1, 0b1, 0b1, 0b1, 0b0, 0b0, 0b0, 0b1, 0b0, 0b0, 0b0, 0b0, 0b1, 0b0, 0b1, 0b1, 0b1, 0b1, 0b1},
		{0b1, 0b1, 0b0, 0b0, 0b0, 0b0, 0b1, 0b1, 0b1, 0b0, 0b0, 0b0, 0b0, 0b1, 0b1, 0b0, 0b1, 0b0, 0b1, 0b0, 0b1, 0b1, 0b1, 0b0, 0b0, 0b0, 0b1, 0b0, 0b1, 0b1, 0b1, 0b0, 0b1, 0b0, 0b0, 0b1, 0b1, 0b0, 0b0, 0b1, 0b1, 0b0, 0b0, 0b1, 0b1, 0b0, 0b1, 0b0, 0b1, 0b1, 0b0, 0b0, 0b0, 0b0, 0b1, 0b1, 0b1, 0b0, 0b1, 0b1, 0b1, 0b0, 0b1, 0b1, 0b1, 0b0, 0b1, 0b1, 0b1, 0b1, 0b0, 0b1, 0b1, 0b1, 0b1, 0b1, 0b0, 0b0, 0b0, 0b1, 0b0, 0b1, 0b0, 0b0, 0b1, 0b1, 0b1, 0b0, 0b1, 0b1, 0b0, 0b1, 0b1, 0b0, 0b1, 0b0, 0b0, 0b0, 0b1, 0b0, 0b0, 0b0, 0b1, 0b0, 0b1, 0b1, 0b0, 0b1, 0b1, 0b1, 0b1, 0b0, 0b0, 0b1, 0b0, 0b0, 0b1, 0b1, 0b1, 0b1, 0b1, 0b0, 0b1, 0b1, 0b0, 0b0, 0b0, 0b1, 0b0, 0b0, 0b1, 0b1, 0b0, 0b1, 0b1, 0b1, 0b1, 0b0, 0b0, 0b1, 0b0, 0b0, 0b0, 0b0, 0b0, 0b1, 0b1, 0b0, 0b0, 0b0, 0b0, 0b0, 0b1, 0b1, 0b0, 0b0, 0b1, 0b0, 0b1, 0b0, 0b0, 0b1, 0b1, 0b1, 0b0, 0b1, 0b0, 0b0, 0b1, 0b0, 0b0, 0b0, 0b1, 0b0, 0b1, 0b1, 0b0, 0b1, 0b0, 0b0, 0b1, 0b0, 0b1, 0b1, 0b1, 0b1, 0b0, 0b0, 0b1, 0b1, 0b0, 0b0, 0b1, 0b1, 0b0, 0b0, 0b0, 0b1, 0b1, 0b0, 0b1, 0b1, 0b1, 0b0, 0b1, 0b1, 0b1, 0b0, 0b0, 0b1, 0b1, 0b0, 0b0, 0b1, 0b0, 0b1, 0b0, 0b0, 0b1, 0b0, 0b0, 0b1, 0b1, 0b0, 0b0, 0b0, 0b1, 0b0, 0b0, 0b1, 0b1, 0b0, 0b1, 0b0, 0b0, 0b1, 0b1, 0b1, 0b1, 0b1, 0b1, 0b0, 0b0, 0b0, 0b1, 0b0, 0b0, 0b0, 0b0, 0b1, 0b0, 0b1, 0b1, 0b1, 0b1, 0b0}}
	valHashes := [][]byte{{0b0, 0b0}, {0b0, 0b1}}
	setTree, _ := NewSetTree(prefixes, valHashes)

	CheckSetTreeConstruction(t, setTree, prefixes, valHashes)
}

func TestAnotherFullHashLengthPrefixes(t *testing.T) {
	prefixes := [][]byte{{0b1, 0b1, 0b0, 0b1, 0b0, 0b1, 0b0, 0b0, 0b1, 0b0, 0b1, 0b1, 0b0, 0b0, 0b1, 0b1, 0b1, 0b0, 0b0, 0b1, 0b1, 0b1, 0b1, 0b1, 0b1, 0b0, 0b1, 0b0, 0b1, 0b1, 0b1, 0b1, 0b1, 0b1, 0b0, 0b1, 0b1, 0b0, 0b1, 0b1, 0b0, 0b0, 0b0, 0b0, 0b1, 0b0, 0b0, 0b1, 0b1, 0b1, 0b1, 0b1, 0b0, 0b0, 0b1, 0b0, 0b0, 0b1, 0b0, 0b1, 0b1, 0b0, 0b1, 0b0, 0b1, 0b0, 0b0, 0b1, 0b0, 0b0, 0b0, 0b0, 0b1, 0b1, 0b0, 0b0, 0b1, 0b1, 0b0, 0b0, 0b1, 0b1, 0b1, 0b1, 0b0, 0b0, 0b0, 0b0, 0b1, 0b0, 0b0, 0b0, 0b0, 0b0, 0b1, 0b0, 0b1, 0b1, 0b0, 0b1, 0b1, 0b1, 0b1, 0b0, 0b0, 0b1, 0b1, 0b1, 0b0, 0b1, 0b1, 0b1, 0b0, 0b0, 0b0, 0b1, 0b1, 0b0, 0b1, 0b0, 0b0, 0b0, 0b0, 0b0, 0b1, 0b0, 0b0, 0b1, 0b0, 0b0, 0b0, 0b0, 0b0, 0b0, 0b0, 0b0, 0b1, 0b0, 0b1, 0b1, 0b1, 0b0, 0b0, 0b0, 0b1, 0b0, 0b1, 0b0, 0b0, 0b1, 0b0, 0b0, 0b0, 0b1, 0b0, 0b0, 0b0, 0b1, 0b0, 0b1, 0b0, 0b0, 0b0, 0b1, 0b1, 0b1, 0b1, 0b0, 0b0, 0b0, 0b1, 0b0, 0b1, 0b0, 0b1, 0b0, 0b0, 0b1, 0b0, 0b1, 0b1, 0b0, 0b1, 0b1, 0b0, 0b0, 0b0, 0b1, 0b0, 0b0, 0b0, 0b1, 0b1, 0b1, 0b1, 0b0, 0b0, 0b1, 0b0, 0b0, 0b0, 0b0, 0b0, 0b0, 0b1, 0b1, 0b1, 0b1, 0b1, 0b0, 0b1, 0b1, 0b0, 0b0, 0b1, 0b0, 0b0, 0b1, 0b0, 0b0, 0b1, 0b1, 0b1, 0b0, 0b1, 0b0, 0b0, 0b0, 0b1, 0b0, 0b1, 0b0, 0b0, 0b1, 0b1, 0b0, 0b1, 0b1, 0b1, 0b1, 0b0, 0b0, 0b0, 0b0, 0b1, 0b0, 0b1, 0b0, 0b0, 0b0, 0b0, 0b0, 0b1, 0b1, 0b0, 0b1}, {0b0, 0b0, 0b1, 0b1, 0b1, 0b1, 0b0, 0b1, 0b0, 0b1, 0b1, 0b0, 0b1, 0b0, 0b0, 0b0, 0b1, 0b1, 0b1, 0b1, 0b0, 0b0, 0b1, 0b0, 0b0, 0b0, 0b0, 0b1, 0b1, 0b0, 0b0, 0b1, 0b0, 0b1, 0b1, 0b1, 0b1, 0b0, 0b1, 0b1, 0b0, 0b1, 0b1, 0b0, 0b1, 0b1, 0b1, 0b1, 0b0, 0b0, 0b1, 0b0, 0b0, 0b0, 0b0, 0b0, 0b1, 0b1, 0b1, 0b1, 0b1, 0b1, 0b0, 0b0, 0b0, 0b1, 0b1, 0b0, 0b1, 0b1, 0b1, 0b1, 0b1, 0b1, 0b0, 0b0, 0b0, 0b0, 0b0, 0b1, 0b1, 0b1, 0b1, 0b0, 0b0, 0b0, 0b0, 0b0, 0b1, 0b1, 0b0, 0b1, 0b0, 0b1, 0b0, 0b0, 0b0, 0b1, 0b0, 0b1, 0b0, 0b0, 0b0, 0b0, 0b0, 0b0, 0b0, 0b1, 0b1, 0b1, 0b1, 0b0, 0b0, 0b0, 0b0, 0b0, 0b1, 0b1, 0b1, 0b1, 0b0, 0b1, 0b0, 0b1, 0b0, 0b0, 0b1, 0b0, 0b1, 0b1, 0b1, 0b1, 0b0, 0b1, 0b0, 0b1, 0b1, 0b0, 0b0, 0b0, 0b0, 0b1, 0b0, 0b0, 0b1, 0b0, 0b1, 0b0, 0b0, 0b1, 0b0, 0b1, 0b0, 0b1, 0b0, 0b1, 0b0, 0b0, 0b1, 0b0, 0b0, 0b1, 0b1, 0b1, 0b0, 0b1, 0b1, 0b1, 0b0, 0b1, 0b1, 0b0, 0b1, 0b1, 0b1, 0b1, 0b1, 0b1, 0b1, 0b0, 0b1, 0b0, 0b1, 0b1, 0b0, 0b1, 0b0, 0b1, 0b1, 0b0, 0b1, 0b0, 0b0, 0b1, 0b0, 0b1, 0b0, 0b1, 0b1, 0b0, 0b0, 0b0, 0b0, 0b0, 0b0, 0b1, 0b0, 0b1, 0b0, 0b1, 0b1, 0b1, 0b1, 0b1, 0b1, 0b0, 0b1, 0b0, 0b1, 0b1, 0b1, 0b1, 0b1, 0b1, 0b1, 0b0, 0b1, 0b0, 0b1, 0b0, 0b1, 0b1, 0b1, 0b0, 0b0, 0b1, 0b0, 0b1, 0b0, 0b0, 0b0, 0b0, 0b0, 0b0, 0b0, 0b0, 0b1, 0b1, 0b0, 0b0, 0b0, 0b0, 0b0, 0b1, 0b1, 0b1}}
	valHashes := [][]byte{{0b1, 0b0, 0b0, 0b0, 0b1, 0b0, 0b0, 0b1, 0b1, 0b1, 0b1, 0b1, 0b1, 0b0, 0b1, 0b1, 0b0, 0b0, 0b1, 0b1, 0b1, 0b0, 0b1, 0b0, 0b0, 0b1, 0b1, 0b0, 0b1, 0b0, 0b1, 0b0, 0b0, 0b1, 0b1, 0b1, 0b1, 0b1, 0b0, 0b0, 0b1, 0b0, 0b1, 0b1, 0b0, 0b0, 0b1, 0b0, 0b0, 0b0, 0b1, 0b1, 0b0, 0b1, 0b0, 0b1, 0b1, 0b0, 0b1, 0b0, 0b1, 0b1, 0b1, 0b0, 0b1, 0b0, 0b1, 0b1, 0b0, 0b1, 0b0, 0b0, 0b0, 0b1, 0b0, 0b1, 0b1, 0b0, 0b0, 0b1, 0b1, 0b0, 0b0, 0b0, 0b1, 0b1, 0b0, 0b1, 0b1, 0b0, 0b0, 0b0, 0b0, 0b1, 0b1, 0b0, 0b1, 0b0, 0b0, 0b1, 0b0, 0b0, 0b0, 0b1, 0b0, 0b0, 0b1, 0b0, 0b1, 0b0, 0b1, 0b1, 0b1, 0b1, 0b0, 0b0, 0b1, 0b0, 0b0, 0b0, 0b0, 0b0, 0b1, 0b0, 0b1, 0b0, 0b0, 0b0, 0b0, 0b1, 0b1, 0b1, 0b0, 0b1, 0b0, 0b0, 0b0, 0b0, 0b0, 0b0, 0b1, 0b0, 0b0, 0b0, 0b1, 0b0, 0b1, 0b0, 0b0, 0b0, 0b1, 0b1, 0b1, 0b0, 0b1, 0b1, 0b0, 0b1, 0b1, 0b1, 0b1, 0b0, 0b0, 0b0, 0b1, 0b0, 0b0, 0b0, 0b1, 0b0, 0b0, 0b0, 0b0, 0b0, 0b0, 0b0, 0b0, 0b1, 0b1, 0b1, 0b1, 0b1, 0b0, 0b0, 0b1, 0b1, 0b0, 0b0, 0b0, 0b1, 0b1, 0b1, 0b1, 0b1, 0b1, 0b0, 0b0, 0b1, 0b0, 0b1, 0b1, 0b1, 0b0, 0b0, 0b0, 0b1, 0b1, 0b0, 0b1, 0b1, 0b0, 0b0, 0b1, 0b0, 0b1, 0b1, 0b0, 0b1, 0b0, 0b0, 0b0, 0b0, 0b1, 0b0, 0b0, 0b0, 0b0, 0b1, 0b1, 0b0, 0b0, 0b1, 0b1, 0b1, 0b0, 0b1, 0b1, 0b0, 0b1, 0b1, 0b0, 0b0, 0b1, 0b0, 0b1, 0b1, 0b0, 0b0, 0b1, 0b0, 0b1, 0b0, 0b1, 0b0, 0b0, 0b1}, {0b0, 0b0, 0b0, 0b0, 0b1, 0b0, 0b0, 0b0, 0b0, 0b0, 0b0, 0b1, 0b1, 0b1, 0b0, 0b1, 0b0, 0b0, 0b1, 0b1, 0b0, 0b1, 0b0, 0b0, 0b1, 0b1, 0b0, 0b1, 0b1, 0b0, 0b0, 0b1, 0b0, 0b0, 0b1, 0b1, 0b0, 0b1, 0b0, 0b1, 0b0, 0b0, 0b0, 0b1, 0b1, 0b0, 0b1, 0b0, 0b0, 0b1, 0b1, 0b0, 0b0, 0b1, 0b1, 0b0, 0b0, 0b1, 0b1, 0b1, 0b0, 0b0, 0b0, 0b1, 0b0, 0b0, 0b0, 0b0, 0b1, 0b1, 0b1, 0b0, 0b1, 0b1, 0b1, 0b1, 0b1, 0b1, 0b0, 0b0, 0b1, 0b1, 0b1, 0b0, 0b0, 0b0, 0b1, 0b0, 0b0, 0b0, 0b0, 0b0, 0b1, 0b0, 0b1, 0b1, 0b1, 0b1, 0b0, 0b0, 0b0, 0b1, 0b1, 0b1, 0b0, 0b0, 0b1, 0b0, 0b0, 0b1, 0b0, 0b0, 0b1, 0b1, 0b0, 0b1, 0b0, 0b0, 0b0, 0b0, 0b1, 0b0, 0b1, 0b0, 0b1, 0b1, 0b0, 0b1, 0b1, 0b0, 0b1, 0b1, 0b0, 0b1, 0b0, 0b0, 0b0, 0b1, 0b1, 0b1, 0b1, 0b0, 0b1, 0b1, 0b1, 0b1, 0b0, 0b0, 0b0, 0b0, 0b0, 0b1, 0b0, 0b0, 0b0, 0b1, 0b1, 0b1, 0b0, 0b0, 0b0, 0b0, 0b0, 0b1, 0b1, 0b1, 0b1, 0b1, 0b0, 0b0, 0b0, 0b0, 0b0, 0b1, 0b0, 0b0, 0b0, 0b1, 0b0, 0b0, 0b1, 0b1, 0b1, 0b0, 0b1, 0b0, 0b1, 0b1, 0b1, 0b0, 0b0, 0b0, 0b0, 0b1, 0b0, 0b0, 0b0, 0b0, 0b0, 0b0, 0b0, 0b0, 0b0, 0b1, 0b0, 0b1, 0b1, 0b1, 0b0, 0b0, 0b1, 0b1, 0b0, 0b1, 0b0, 0b0, 0b0, 0b0, 0b0, 0b1, 0b0, 0b1, 0b0, 0b1, 0b1, 0b0, 0b1, 0b1, 0b0, 0b0, 0b1, 0b0, 0b0, 0b1, 0b1, 0b1, 0b0, 0b0, 0b1, 0b0, 0b0, 0b0, 0b1, 0b1, 0b0, 0b0, 0b1, 0b1, 0b0, 0b0, 0b0, 0b1, 0b0, 0b1, 0b0, 0b0}}
	setTree, _ := NewSetTree(prefixes, valHashes)

	CheckSetTreeConstruction(t, setTree, prefixes, valHashes)
}

// TODO: fix the test below
/* func TestSingleEntry(t *testing.T) {
	prefixes := [][]byte{{0b1, 0b0}}
	valHashes := [][]byte{{0b0, 0b0}}
	setTree, _ := NewSetTree(prefixes, valHashes)

	CheckSetTreeConstruction(t, setTree, prefixes, valHashes)
} */

/* Not a test, just a helper function */
func CheckSetTreeConstruction(t *testing.T, setTree *SetTree, prefixes [][]byte, valHashes [][]byte) {
	for _, prefix := range prefixes {
		if !setTree.HasKey(prefix) {
			t.Errorf("Set tree does not have key %b", prefix)
		}
	}

	for i, prefix := range prefixes {
		if !setTree.HasKey(prefix) {
			t.Errorf("Set tree does not have key %b", prefix)
			continue
		}
		found := false
		for _, value := range setTree.GetLeaf(prefix).getValues() {
			if bytes.Equal(value.Hash, valHashes[i]) {
				found = true
			}
		}
		if !found {
			t.Error("Unable to find value for prefix ", prefix)
		}
	}
}

func TestNonexistence(t *testing.T) {
	/*
				r
			  0    111
		   0
		 0   1
	*/

	/*

				r
			0      1
		   0        1
		  0          1
	*/
	prefixes := [][]byte{{0b0, 0b0, 0b0}, {0b1, 0b1, 0b1}, {0b0, 0b0, 0b1}}
	valHashes := [][]byte{{0b0, 0b0, 0b0}, {0b0, 0b0, 0b01}, {0b0, 0b1, 0b0}}
	setTree, _ := NewSetTree(prefixes, valHashes)

	CheckSetTreeConstruction(t, setTree, prefixes, valHashes)

	prefix := []byte{0b0, 0b1, 0b1}

	proof, leafValues := setTree.ProveNonExistence(prefix)
	_ = leafValues
	if proof == nil {
		t.Errorf("Prove for prefix %b was nil", prefix)
	}
}
