package eip3009

import (
	"errors"
	"math/big"

	"github.com/gosunuts/ethtxbuilder/utils"
	"github.com/umbracle/ethgo"
	"github.com/umbracle/ethgo/abi"
)

const RawABI = `[
  {
    "name": "transferWithAuthorization",
    "type": "function",
    "inputs": [
      { "name": "from", "type": "address" },
      { "name": "to", "type": "address" },
      { "name": "value", "type": "uint256" },
      { "name": "validAfter", "type": "uint256" },
      { "name": "validBefore", "type": "uint256" },
      { "name": "nonce", "type": "bytes32" },
      { "name": "signature", "type": "bytes" }
    ],
    "outputs": [],
    "stateMutability": "nonpayable"
  },
  {
    "name": "balanceOf",
    "type": "function",
    "inputs": [
      { "name": "account", "type": "address" }
    ],
    "outputs": [
      { "name": "balance", "type": "uint256" }
    ],
    "stateMutability": "view"
  }
]`

type Runtime struct {
	a *abi.ABI
}

func New() (*Runtime, error) {
	a, err := abi.NewABI(RawABI)
	if err != nil {
		return nil, err
	}
	return &Runtime{a: a}, nil
}

// ------------------------- Pack (tx data) ------------------------------------

// PackTransferWithAuth builds calldata for transferWithAuthorization(...).
func (r *Runtime) PackTransferWithAuth(
	from ethgo.Address,
	to ethgo.Address,
	value *big.Int,
	validAfter *big.Int,
	validBefore *big.Int,
	nonce ethgo.Hash,
	signature []byte,
) ([]byte, error) {
	m := r.a.Methods["transferWithAuthorization"]
	if m == nil {
		return nil, errors.New("method not found: transferWithAuthorization")
	}
	return m.Encode([]interface{}{
		from, to, value, validAfter, validBefore, nonce, signature,
	})
}

// PackBalanceOf builds calldata for balanceOf(account).
func (r *Runtime) PackBalanceOf(account ethgo.Address) ([]byte, error) {
	m := r.a.Methods["balanceOf"]
	if m == nil {
		return nil, errors.New("method not found: balanceOf")
	}
	return m.Encode([]interface{}{account})
}

// ------------------------- Decode (call outputs) ------------------------------

// DecodeBalanceOfOutput decodes eth_call hex output (e.g. "0x...") to *big.Int.
func (r *Runtime) DecodeBalanceOfOutput(outputHex string) (*big.Int, error) {
	b, err := utils.HexToBytes(outputHex)
	if err != nil {
		return nil, err
	}
	m := r.a.Methods["balanceOf"]
	if m == nil {
		return nil, errors.New("method not found: balanceOf")
	}
	out, err := m.Decode(b)
	if err != nil {
		return nil, err
	}
	if len(out) != 1 {
		return nil, errors.New("unexpected outputs length")
	}
	bal, ok := out["balance"].(*big.Int)
	if !ok {
		return nil, errors.New("unexpected output type (want *big.Int)")
	}
	return bal, nil
}
