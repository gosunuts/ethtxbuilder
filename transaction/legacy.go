package transaction

import (
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/gosunuts/ethtxbuilder/utils"
	"github.com/umbracle/fastrlp"
)

func NewLegacyTx(nonce uint64, to string, amount *big.Int, gasLimit uint64, gasPrice *big.Int, data []byte) *LegacyTx {
	return &LegacyTx{
		Nonce:    nonce,
		GasPrice: gasPrice,
		Gas:      gasLimit,
		To:       utils.StrToRawAddr(to),
		Value:    amount,
		Data:     data,
	}
}

type LegacyTx struct {
	Nonce    uint64
	GasPrice *big.Int
	Gas      uint64
	To       []byte
	Value    *big.Int
	Data     []byte
	ChainID  *big.Int

	V, R, S *big.Int

	rawtx []byte
}

func (t *LegacyTx) Sender() (string, error) {
	sigHash := utils.Keccak(t.sigPayloadRLP())
	v := t.V
	if t.ChainID != nil && t.ChainID.Sign() > 0 {
		v = new(big.Int).Sub(v, new(big.Int).Mul(t.ChainID, big.NewInt(2)))
		v.Sub(v, big.NewInt(8))
	}
	return utils.RecoverFrom(sigHash, t.R, t.S, v, true)
}

func (t *LegacyTx) EncodeRLP() []byte {
	if t.rawtx != nil {
		return t.rawtx
	}

	var ar fastrlp.Arena
	l := ar.NewArray()
	l.Set(ar.NewUint(t.Nonce))
	l.Set(utils.SetBigOrZero(&ar, t.GasPrice))
	l.Set(ar.NewUint(t.Gas))
	l.Set(utils.SetTo(&ar, t.To))
	l.Set(utils.SetBigOrZero(&ar, t.Value))
	l.Set(ar.NewBytes(t.Data))

	l.Set(utils.SetBigOrZero(&ar, t.V))
	l.Set(utils.SetBigOrZero(&ar, t.R))
	l.Set(utils.SetBigOrZero(&ar, t.S))
	t.rawtx = append([]byte{}, l.MarshalTo(nil)...)
	return t.rawtx
}

func (t *LegacyTx) Sign(sign utils.SignFunc) error {
	preimage := t.sigPayloadRLP()
	msgHash := utils.Keccak(preimage)

	// 65 bytes: r(32) || s(32) || yParity(1) where yParity in {0,1}
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

	// calc v
	if t.ChainID != nil && t.ChainID.Sign() > 0 {
		// EIP-155: v = 35 + 2*chainId + yParity
		v := new(big.Int).Mul(t.ChainID, big.NewInt(2))
		v.Add(v, big.NewInt(int64(35+y)))
		t.V, t.R, t.S = v, r, s
		return nil
	}

	// Unprotected legacy: v = 27 + yParity
	v := new(big.Int).SetUint64(27 + y)
	t.V, t.R, t.S = v, r, s
	return nil
}

func (t *LegacyTx) sigPayloadRLP() []byte {
	var ar fastrlp.Arena
	l := ar.NewArray()
	l.Set(ar.NewUint(t.Nonce))
	l.Set(utils.SetBigOrZero(&ar, t.GasPrice))
	l.Set(ar.NewUint(t.Gas))
	l.Set(utils.SetTo(&ar, t.To))
	l.Set(utils.SetBigOrZero(&ar, t.Value))
	l.Set(ar.NewBytes(t.Data))

	if t.ChainID != nil && t.ChainID.Sign() > 0 {
		l.Set(ar.NewBigInt(t.ChainID))
		l.Set(ar.NewUint(0))
		l.Set(ar.NewUint(0))
	}
	return l.MarshalTo(nil)
}

func (t *LegacyTx) TxHash() string {
	raw := t.EncodeRLP()
	hash := utils.Keccak(raw)
	return "0x" + hex.EncodeToString(hash)
}
