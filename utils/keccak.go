package utils

import "golang.org/x/crypto/sha3"

func Keccak(b []byte) []byte {
	h := sha3.NewLegacyKeccak256()
	h.Write(b)
	return h.Sum(nil)
}
