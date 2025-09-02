package transaction

import (
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/gosunuts/ethtxbuilder/utils"
	"github.com/umbracle/fastrlp"
)

type AccessTuple struct {
	Address     []byte   // 20 bytes
	StorageKeys [][]byte // 32-byte slots
}
type AccessList []AccessTuple

func setAccessList(ar *fastrlp.Arena, al AccessList) *fastrlp.Value {
	l := ar.NewArray()
	for _, t := range al {
		it := ar.NewArray()
		it.Set(utils.SetTo(ar, t.Address))
		slots := ar.NewArray()
		for _, k := range t.StorageKeys {
			slots.Set(ar.NewBytes(k))
		}
		it.Set(slots)
		l.Set(it)
	}
	return l
}

func NewDynamicTx(chainId *big.Int, nonce uint64, to string, amount *big.Int, gasLimit uint64, maxPriorityFeePerGas, maxFeePerGas *big.Int, data []byte) *DynamicTx {
	return &DynamicTx{
		ChainID:              chainId,
		Nonce:                nonce,
		MaxPriorityFeePerGas: maxPriorityFeePerGas,
		MaxFeePerGas:         maxFeePerGas,
		Gas:                  gasLimit,
		To:                   utils.StrToRawAddr(to),
		Value:                amount,
		Data:                 data,
	}
}

type DynamicTx struct {
	ChainID              *big.Int
	Nonce                uint64
	MaxPriorityFeePerGas *big.Int
	MaxFeePerGas         *big.Int
	Gas                  uint64
	To                   []byte // 20 bytes or nil for creation
	Value                *big.Int
	Data                 []byte
	Accesses             AccessList

	V, R, S *big.Int

	rawtx []byte
}

func (t *DynamicTx) sigPayloadRLP() []byte {
	var ar fastrlp.Arena
	l := ar.NewArray()
	l.Set(ar.NewBigInt(t.ChainID))
	l.Set(ar.NewUint(t.Nonce))
	l.Set(utils.SetBigOrZero(&ar, t.MaxPriorityFeePerGas))
	l.Set(utils.SetBigOrZero(&ar, t.MaxFeePerGas))
	l.Set(ar.NewUint(t.Gas))
	l.Set(utils.SetTo(&ar, t.To))
	l.Set(utils.SetBigOrZero(&ar, t.Value))
	l.Set(ar.NewBytes(t.Data))
	l.Set(setAccessList(&ar, t.Accesses))

	payload := l.MarshalTo(nil)
	return append([]byte{utils.DynamicFeeTxType}, payload...)
}

func (t *DynamicTx) EncodeRLP() []byte {
	if t.rawtx != nil {
		return t.rawtx
	}
	var ar fastrlp.Arena
	l := ar.NewArray()
	l.Set(ar.NewBigInt(t.ChainID))
	l.Set(ar.NewUint(t.Nonce))
	l.Set(utils.SetBigOrZero(&ar, t.MaxPriorityFeePerGas))
	l.Set(utils.SetBigOrZero(&ar, t.MaxFeePerGas))
	l.Set(ar.NewUint(t.Gas))
	l.Set(utils.SetTo(&ar, t.To))
	l.Set(utils.SetBigOrZero(&ar, t.Value))
	l.Set(ar.NewBytes(t.Data))
	l.Set(setAccessList(&ar, t.Accesses))
	l.Set(utils.SetBigOrZero(&ar, t.V))
	l.Set(utils.SetBigOrZero(&ar, t.R))
	l.Set(utils.SetBigOrZero(&ar, t.S))

	payload := l.MarshalTo(nil)
	t.rawtx = append([]byte{utils.DynamicFeeTxType}, payload...)
	return t.rawtx
}

func (t *DynamicTx) Sender() (string, error) {
	if t.ChainID == nil || t.ChainID.Sign() <= 0 {
		return "", fmt.Errorf("chainID required for 1559 sender recovery")
	}
	sighash := utils.Keccak(t.sigPayloadRLP())
	v27 := new(big.Int).Add(t.V, big.NewInt(27)) // 0/1 -> 27/28
	return utils.RecoverFrom(sighash, t.R, t.S, v27, true)
}

func (t *DynamicTx) Sign(sign utils.SignFunc) error {
	if t.ChainID == nil || t.ChainID.Sign() <= 0 {
		return fmt.Errorf("chainID required for 1559")
	}
	preimage := t.sigPayloadRLP()
	msgHash := utils.Keccak(preimage)

	sig, err := sign(msgHash)
	if err != nil {
		return err
	}
	if len(sig) != 65 {
		return fmt.Errorf("signature must be 65 bytes, got %d", len(sig))
	}

	r := new(big.Int).SetBytes(sig[:32])
	s := new(big.Int).SetBytes(sig[32:64])
	y := uint64(sig[64] & 0x01)

	t.V = new(big.Int).SetUint64(y) // 0/1
	t.R, t.S = r, s

	t.rawtx = nil
	_ = t.EncodeRLP()
	return nil
}

func (t *DynamicTx) TxHash() string {
	raw := t.EncodeRLP()
	hash := utils.Keccak(raw)
	return "0x" + hex.EncodeToString(hash)
}
