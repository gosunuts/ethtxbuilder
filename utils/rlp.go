package utils

import (
	"math/big"

	"github.com/umbracle/fastrlp"
)

func SetBigOrZero(ar *fastrlp.Arena, x *big.Int) *fastrlp.Value {
	if x == nil || x.Sign() == 0 {
		return ar.NewUint(0)
	}
	return ar.NewBigInt(x)
}

func SetTo(ar *fastrlp.Arena, to []byte) *fastrlp.Value {
	if len(to) == 0 {
		return ar.NewBytes([]byte{}) // empty → contract creation
	}
	return ar.NewBytes(to) // 20 bytes
}
