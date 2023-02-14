package utils

import "github.com/minio/blake2b-simd"

func Blake256(data []byte) []byte {
	hasher := blake2b.New256()
	hasher.Write(data)
	return hasher.Sum(nil)
}
