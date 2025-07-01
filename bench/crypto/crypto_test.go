package bench

import (
	"testing"

	libcrypto "github.com/huyuncong/MerkleSquare/lib/crypto"
	"github.com/immesys/bw2/crypto"
)

func helperHash(b *testing.B) {
	content := make([]byte, 256)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		libcrypto.Hash(content)
	}
}

func helperSign(b *testing.B) {
	sk, vk := crypto.GenerateKeypair()
	signature := make([]byte, 64)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		crypto.SignBlob(sk, vk, signature, vk)
	}
}

func helperVerify(b *testing.B) {
	sk, vk := crypto.GenerateKeypair()
	signature := make([]byte, 64)
	crypto.SignBlob(sk, vk, signature, vk)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		crypto.VerifyBlob(vk, signature, vk)
	}
}

func BenchmarkHash(b *testing.B) {
	helperHash(b)
}

func BenchmarkSign(b *testing.B) {
	helperSign(b)
}

func BenchmarkVerify(b *testing.B) {
	helperVerify(b)
}
