package utils

import (
	"github.com/umbracle/ethgo"
)

func PubkeyToAddr(pubkey []byte) []byte {
	return Keccak(pubkey[1:])[12:]
}

func StrToRawAddr(s string) []byte {
	addr := ethgo.HexToAddress(s)
	return addr[:]
}

func RawAddrToStr(addr []byte) string {
	a := ethgo.BytesToAddress(addr)
	return a.String()
}
