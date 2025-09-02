package erc20

import (
	"bytes"
	"errors"
	"math/big"

	"github.com/gosunuts/ethtxbuilder/utils"
	"github.com/umbracle/ethgo"
	"github.com/umbracle/ethgo/abi"
)

const RawABI = `[
  {"name":"name","type":"function","stateMutability":"view","inputs":[],"outputs":[{"name":"","type":"string"}]},
  {"name":"symbol","type":"function","stateMutability":"view","inputs":[],"outputs":[{"name":"","type":"string"}]},
  {"name":"decimals","type":"function","stateMutability":"view","inputs":[],"outputs":[{"name":"","type":"uint8"}]},
  {"name":"totalSupply","type":"function","stateMutability":"view","inputs":[],"outputs":[{"name":"","type":"uint256"}]},
  {"name":"balanceOf","type":"function","stateMutability":"view","inputs":[{"name":"account","type":"address"}],"outputs":[{"name":"balance","type":"uint256"}]},
  {"name":"allowance","type":"function","stateMutability":"view","inputs":[{"name":"owner","type":"address"},{"name":"spender","type":"address"}],"outputs":[{"name":"remaining","type":"uint256"}]},
  {"name":"transfer","type":"function","stateMutability":"nonpayable","inputs":[{"name":"to","type":"address"},{"name":"value","type":"uint256"}],"outputs":[{"name":"","type":"bool"}]},
  {"name":"approve","type":"function","stateMutability":"nonpayable","inputs":[{"name":"spender","type":"address"},{"name":"value","type":"uint256"}],"outputs":[{"name":"","type":"bool"}]},
  {"name":"transferFrom","type":"function","stateMutability":"nonpayable","inputs":[{"name":"from","type":"address"},{"name":"to","type":"address"},{"name":"value","type":"uint256"}],"outputs":[{"name":"","type":"bool"}]},
  {"anonymous":false,"name":"Transfer","type":"event","inputs":[
    {"indexed":true,"name":"from","type":"address"},
    {"indexed":true,"name":"to","type":"address"},
    {"indexed":false,"name":"value","type":"uint256"}]},
  {"anonymous":false,"name":"Approval","type":"event","inputs":[
    {"indexed":true,"name":"owner","type":"address"},
    {"indexed":true,"name":"spender","type":"address"},
    {"indexed":false,"name":"value","type":"uint256"}]}
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

/* ----------------------------- Pack (tx data) ------------------------------ */

func (r *Runtime) PackTransfer(to ethgo.Address, value *big.Int) ([]byte, error) {
	m := r.a.Methods["transfer"]
	if m == nil {
		return nil, errors.New("method not found: transfer")
	}
	return m.Encode([]any{to, value})
}

func (r *Runtime) PackApprove(spender ethgo.Address, value *big.Int) ([]byte, error) {
	m := r.a.Methods["approve"]
	if m == nil {
		return nil, errors.New("method not found: approve")
	}
	return m.Encode([]any{spender, value})
}

func (r *Runtime) PackTransferFrom(from, to ethgo.Address, value *big.Int) ([]byte, error) {
	m := r.a.Methods["transferFrom"]
	if m == nil {
		return nil, errors.New("method not found: transferFrom")
	}
	return m.Encode([]any{from, to, value})
}

func (r *Runtime) PackBalanceOf(acct ethgo.Address) ([]byte, error) {
	m := r.a.Methods["balanceOf"]
	if m == nil {
		return nil, errors.New("method not found: balanceOf")
	}
	return m.Encode([]any{acct})
}

func (r *Runtime) PackAllowance(owner, spender ethgo.Address) ([]byte, error) {
	m := r.a.Methods["allowance"]
	if m == nil {
		return nil, errors.New("method not found: allowance")
	}
	return m.Encode([]any{owner, spender})
}

func (r *Runtime) PackName() ([]byte, error) {
	m := r.a.Methods["name"]
	if m == nil {
		return nil, errors.New("method not found: name")
	}
	return m.Encode(nil)
}

func (r *Runtime) PackSymbol() ([]byte, error) {
	m := r.a.Methods["symbol"]
	if m == nil {
		return nil, errors.New("method not found: symbol")
	}
	return m.Encode(nil)
}

func (r *Runtime) PackDecimals() ([]byte, error) {
	m := r.a.Methods["decimals"]
	if m == nil {
		return nil, errors.New("method not found: decimals")
	}
	return m.Encode(nil)
}

func (r *Runtime) PackTotalSupply() ([]byte, error) {
	m := r.a.Methods["totalSupply"]
	if m == nil {
		return nil, errors.New("method not found: totalSupply")
	}
	return m.Encode(nil)
}

/* ---------------------------- Decode (call outs) --------------------------- */

func (r *Runtime) DecodeUint256Single(outputHex string, method string, preferName string) (*big.Int, error) {
	b, err := utils.HexToBytes(outputHex)
	if err != nil {
		return nil, err
	}
	m := r.a.Methods[method]
	if m == nil {
		return nil, errors.New("method not found: " + method)
	}
	out, err := m.Decode(b)
	if err != nil {
		return nil, err
	}

	if v, ok := out[preferName]; ok {
		if z, ok := v.(*big.Int); ok {
			return z, nil
		}
		return nil, errors.New("unexpected type for " + preferName)
	}
	if v, ok := out["0"]; ok {
		if z, ok := v.(*big.Int); ok {
			return z, nil
		}
	}
	return nil, errors.New("uint256 output not found")
}

func (r *Runtime) DecodeStringSingle(outputHex string, method string) (string, error) {
	b, err := utils.HexToBytes(outputHex)
	if err != nil {
		return "", err
	}
	m := r.a.Methods[method]
	if m == nil {
		return "", errors.New("method not found: " + method)
	}
	out, err := m.Decode(b)
	if err != nil {
		return "", err
	}

	if v, ok := out["0"]; ok {
		if s, ok := v.(string); ok {
			return s, nil
		}
	}
	for _, key := range []string{"name", "symbol"} {
		if v, ok := out[key]; ok {
			if s, ok := v.(string); ok {
				return s, nil
			}
		}
	}
	return "", errors.New("string output not found")
}

func (r *Runtime) DecodeUint8Single(outputHex string, method string) (uint8, error) {
	b, err := utils.HexToBytes(outputHex)
	if err != nil {
		return 0, err
	}
	m := r.a.Methods[method]
	if m == nil {
		return 0, errors.New("method not found: " + method)
	}
	out, err := m.Decode(b)
	if err != nil {
		return 0, err
	}

	if v, ok := out["0"]; ok {
		if bi, ok := v.(*big.Int); ok {
			return uint8(bi.Uint64()), nil
		}
		if u8, ok := v.(uint8); ok {
			return u8, nil
		}
	}
	if v, ok := out["decimals"]; ok {
		if bi, ok := v.(*big.Int); ok {
			return uint8(bi.Uint64()), nil
		}
		if u8, ok := v.(uint8); ok {
			return u8, nil
		}
	}
	return 0, errors.New("uint8 output not found")
}

/* ---- Convenience wrappers for common reads ---- */

func (r *Runtime) DecodeBalanceOf(outputHex string) (*big.Int, error) {
	return r.DecodeUint256Single(outputHex, "balanceOf", "balance")
}
func (r *Runtime) DecodeAllowance(outputHex string) (*big.Int, error) {
	return r.DecodeUint256Single(outputHex, "allowance", "remaining")
}
func (r *Runtime) DecodeTotalSupply(outputHex string) (*big.Int, error) {
	return r.DecodeUint256Single(outputHex, "totalSupply", "0")
}
func (r *Runtime) DecodeName(outputHex string) (string, error) {
	return r.DecodeStringSingle(outputHex, "name")
}
func (r *Runtime) DecodeSymbol(outputHex string) (string, error) {
	return r.DecodeStringSingle(outputHex, "symbol")
}
func (r *Runtime) DecodeDecimals(outputHex string) (uint8, error) {
	return r.DecodeUint8Single(outputHex, "decimals")
}

/* ---------------------------- Events (logs) -------------------------------- */

var (
	// keccak256("Transfer(address,address,uint256)")
	transferSig = ethgo.HexToHash("0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef")

	// keccak256("Approval(address,address,uint256)")
	approvalSig = ethgo.HexToHash("0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925")
)

type TransferEvent struct {
	From  ethgo.Address
	To    ethgo.Address
	Value *big.Int
}

type ApprovalEvent struct {
	Owner   ethgo.Address
	Spender ethgo.Address
	Value   *big.Int
}

func (r *Runtime) DecodeTransfer(topics []ethgo.Hash, data []byte) (*TransferEvent, error) {
	// topics[0]=sig, [1]=from, [2]=to ; data = 32-byte value
	if len(topics) < 3 {
		return nil, errors.New("Transfer: need 3 topics")
	}
	if !bytes.Equal(topics[0].Bytes(), transferSig.Bytes()) {
		return nil, errors.New("Transfer: topic[0] mismatch")
	}
	if len(data) < 32 {
		return nil, errors.New("Transfer: data too short")
	}
	from := utils.TopicToAddress(topics[1])
	to := utils.TopicToAddress(topics[2])
	val := new(big.Int).SetBytes(data[len(data)-32:]) // right-aligned 32 bytes
	return &TransferEvent{From: from, To: to, Value: val}, nil
}

func (r *Runtime) DecodeApproval(topics []ethgo.Hash, data []byte) (*ApprovalEvent, error) {
	// topics[0]=sig, [1]=owner, [2]=spender ; data = 32-byte value
	if len(topics) < 3 {
		return nil, errors.New("Approval: need 3 topics")
	}
	if !bytes.Equal(topics[0].Bytes(), approvalSig.Bytes()) {
		return nil, errors.New("Approval: topic[0] mismatch")
	}
	if len(data) < 32 {
		return nil, errors.New("Approval: data too short")
	}
	owner := utils.TopicToAddress(topics[1])
	spender := utils.TopicToAddress(topics[2])
	val := new(big.Int).SetBytes(data[len(data)-32:])
	return &ApprovalEvent{Owner: owner, Spender: spender, Value: val}, nil
}
