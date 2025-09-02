package utils

import (
	"math/big"

	"github.com/umbracle/ethgo"
)

func TopicToAddress(h ethgo.Hash) ethgo.Address {
	var a ethgo.Address
	copy(a[:], h.Bytes()[12:32])
	return a
}

func TopicToUint256(h ethgo.Hash) *big.Int {
	return new(big.Int).SetBytes(h.Bytes()) // big-endian
}
