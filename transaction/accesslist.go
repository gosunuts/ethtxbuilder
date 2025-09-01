package transaction

import (
	"math/big"

	"github.com/gosunuts/ethtxbuilder/utils"
	"github.com/umbracle/fastrlp"
)

type AccessTuple struct {
	Address     []byte   // 20 bytes
	StorageKeys [][]byte // 32 bytes each
}

type AccessListTx struct {
	ChainID    *big.Int
	Nonce      uint64
	GasPrice   *big.Int
	Gas        uint64
	To         []byte // nil => contract creation
	Value      *big.Int
	Data       []byte
	AccessList []AccessTuple

	YParity uint64 // 0 or 1
	R       *big.Int
	S       *big.Int
}

func EncodeAccessList2930(tx *AccessListTx) []byte {
	var ar fastrlp.Arena

	al := ar.NewArray()
	for _, t := range tx.AccessList {
		elem := ar.NewArray()
		elem.Set(ar.NewBytes(t.Address))
		keys := ar.NewArray()
		for _, k := range t.StorageKeys {
			keys.Set(ar.NewBytes(k))
		}
		elem.Set(keys)
		al.Set(elem)
	}

	l := ar.NewArray()
	l.Set(utils.SetBigOrZero(&ar, tx.ChainID))
	l.Set(ar.NewUint(tx.Nonce))
	l.Set(utils.SetBigOrZero(&ar, tx.GasPrice))
	l.Set(ar.NewUint(tx.Gas))
	l.Set(utils.SetTo(&ar, tx.To))
	l.Set(utils.SetBigOrZero(&ar, tx.Value))
	l.Set(ar.NewBytes(tx.Data))
	l.Set(al)
	l.Set(ar.NewUint(tx.YParity))
	l.Set(utils.SetBigOrZero(&ar, tx.R))
	l.Set(utils.SetBigOrZero(&ar, tx.S))

	enc := l.MarshalTo(nil)
	return append([]byte{0x01}, enc...)
}
