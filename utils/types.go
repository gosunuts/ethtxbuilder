package utils

import (
	"fmt"
	"math/big"

	"github.com/umbracle/ethgo"
)

func StrToU64(s string) (uint64, error) {
	if len(s) > 2 && s[:2] == "0x" {
		n := new(big.Int)
		n.SetString(s[2:], 16)
		return n.Uint64(), nil
	}
	n := new(big.Int)
	if _, ok := n.SetString(s, 10); ok {
		return n.Uint64(), nil
	}
	return 0, fmt.Errorf("invalid number format: %s", s)
}

func StrToAddr(s string) ethgo.Address {
	return ethgo.HexToAddress(s)
}
