package main

import (
	"fmt"
	"time"

	crypto "github.com/huyuncong/MerkleSquare/lib/crypto"
)

var result []byte

func BenchmarkHash() int {
	start := time.Now()
	i := 0
	input := []byte("hello world")
	var output []byte
	for time.Since(start) < 10*time.Second {
		output = crypto.Hash(input)
		i++
	}
	result = output

	return i / 10
}

func main() {
	fmt.Println(BenchmarkHash())
}
