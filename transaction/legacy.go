package transaction

import (
	"math/big"

	"github.com/gosunuts/ethtxbuilder/utils"
	"github.com/umbracle/fastrlp"
)

type LegacyTx struct {
	Nonce    uint64
	GasPrice *big.Int
	Gas      uint64
	To       []byte
	Value    *big.Int
	Data     []byte

	V *big.Int
	R *big.Int
	S *big.Int
}

func EncodeLegacy(tx *LegacyTx) []byte {
	var ar fastrlp.Arena
	l := ar.NewArray()
	l.Set(ar.NewUint(tx.Nonce))
	l.Set(utils.SetBigOrZero(&ar, tx.GasPrice))
	l.Set(ar.NewUint(tx.Gas))
	l.Set(utils.SetTo(&ar, tx.To))
	l.Set(utils.SetBigOrZero(&ar, tx.Value))
	l.Set(ar.NewBytes(tx.Data))
	l.Set(utils.SetBigOrZero(&ar, tx.V))
	l.Set(utils.SetBigOrZero(&ar, tx.R))
	l.Set(utils.SetBigOrZero(&ar, tx.S))
	return l.MarshalTo(nil)
}
