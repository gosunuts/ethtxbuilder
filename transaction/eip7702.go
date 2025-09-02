package transaction

import (
	"math/big"

	"github.com/gosunuts/ethtxbuilder/utils"
	"github.com/umbracle/fastrlp"
)

type SetCodeAuthorization struct {
	ChainID *big.Int
	Address []byte // 20 bytes (authorizer)
	Nonce   *big.Int
	YParity uint64
	R       *big.Int
	S       *big.Int
}

type SetCodeTx struct {
	ChainID              *big.Int
	Nonce                uint64
	MaxPriorityFeePerGas *big.Int
	MaxFeePerGas         *big.Int
	Gas                  uint64
	Destination          []byte // EOA/contract; nil=creation
	Value                *big.Int
	Data                 []byte
	AccessList           []AccessTuple

	AuthorizationList []SetCodeAuthorization

	YParity uint64
	R       *big.Int
	S       *big.Int
}

func EncodeSetCode7702(tx *SetCodeTx) []byte {
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

	auths := ar.NewArray()
	for _, a := range tx.AuthorizationList {
		elem := ar.NewArray()
		elem.Set(utils.SetBigOrZero(&ar, a.ChainID))
		elem.Set(ar.NewBytes(a.Address))
		elem.Set(utils.SetBigOrZero(&ar, a.Nonce))
		elem.Set(ar.NewUint(a.YParity))
		elem.Set(utils.SetBigOrZero(&ar, a.R))
		elem.Set(utils.SetBigOrZero(&ar, a.S))
		auths.Set(elem)
	}

	l := ar.NewArray()
	l.Set(utils.SetBigOrZero(&ar, tx.ChainID))
	l.Set(ar.NewUint(tx.Nonce))
	l.Set(utils.SetBigOrZero(&ar, tx.MaxPriorityFeePerGas))
	l.Set(utils.SetBigOrZero(&ar, tx.MaxFeePerGas))
	l.Set(ar.NewUint(tx.Gas))
	l.Set(utils.SetTo(&ar, tx.Destination))
	l.Set(utils.SetBigOrZero(&ar, tx.Value))
	l.Set(ar.NewBytes(tx.Data))
	l.Set(al)
	l.Set(auths)
	l.Set(ar.NewUint(tx.YParity))
	l.Set(utils.SetBigOrZero(&ar, tx.R))
	l.Set(utils.SetBigOrZero(&ar, tx.S))

	enc := l.MarshalTo(nil)
	return append([]byte{0x04}, enc...)
}
