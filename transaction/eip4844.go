package transaction

import (
	"math/big"

	"github.com/gosunuts/ethtxbuilder/utils"
	"github.com/umbracle/fastrlp"
)

type BlobTx struct {
	ChainID              *big.Int
	Nonce                uint64
	MaxPriorityFeePerGas *big.Int
	MaxFeePerGas         *big.Int
	Gas                  uint64
	To                   []byte
	Value                *big.Int
	Data                 []byte
	AccessList           []AccessTuple

	BlobVersionedHashes [][]byte // each 32 bytes (EIP-4844 versioned hash)
	MaxFeePerBlobGas    *big.Int

	YParity uint64
	R       *big.Int
	S       *big.Int
}

func EncodeBlob4844(tx *BlobTx) []byte {
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

	bvh := ar.NewArray()
	for _, h := range tx.BlobVersionedHashes {
		bvh.Set(ar.NewBytes(h))
	}

	l := ar.NewArray()
	l.Set(utils.SetBigOrZero(&ar, tx.ChainID))
	l.Set(ar.NewUint(tx.Nonce))
	l.Set(utils.SetBigOrZero(&ar, tx.MaxPriorityFeePerGas))
	l.Set(utils.SetBigOrZero(&ar, tx.MaxFeePerGas))
	l.Set(ar.NewUint(tx.Gas))
	l.Set(utils.SetTo(&ar, tx.To))
	l.Set(utils.SetBigOrZero(&ar, tx.Value))
	l.Set(ar.NewBytes(tx.Data))
	l.Set(al)
	l.Set(bvh)
	l.Set(utils.SetBigOrZero(&ar, tx.MaxFeePerBlobGas))
	l.Set(ar.NewUint(tx.YParity))
	l.Set(utils.SetBigOrZero(&ar, tx.R))
	l.Set(utils.SetBigOrZero(&ar, tx.S))

	enc := l.MarshalTo(nil)
	return append([]byte{0x03}, enc...)
}
