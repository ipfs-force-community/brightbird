package utils

import "github.com/minio/blake2b-simd"

func Blake256(datas ...[]byte) []byte {
	hasher := blake2b.New256()
	for _, d := range datas {
		hasher.Write(d)
	}
	return hasher.Sum(nil)
}
