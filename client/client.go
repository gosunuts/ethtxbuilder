package client

import (
	"math/big"

	"github.com/gosunuts/ethtxbuilder/utils"
	"github.com/umbracle/ethgo"
	"github.com/umbracle/ethgo/jsonrpc"
)

const (
	EthereumMainnet = "https://ethereum-rpc.publicnode.com"
	EthereumSepolia = "https://ethereum-sepolia-rpc.publicnode.com"
	OptimismMainnet = "https://optimism-rpc.publicnode.com"
	OptimismSepolia = "https://optimism-sepolia-rpc.publicnode.com"
	ArbitrumOne     = "https://arbitrum-one-rpc.publicnode.com"
	ArbitrumSepolia = "https://arbitrum-sepolia-rpc.publicnode.com"
)

type Client struct {
	rpc          *jsonrpc.Client
	ChainId      *big.Int
	NonceManager *NonceManager
}

func NewClient(endpoint string) (*Client, error) {
	c, err := jsonrpc.NewClient(endpoint)
	if err != nil {
		return nil, err
	}
	client := &Client{rpc: c}
	nonceManager := NewNonceManager(client, 0)
	client.NonceManager = nonceManager

	chainID, err := c.Eth().ChainID()
	if err != nil {
		return nil, err
	}
	client.ChainId = chainID

	return client, nil
}

func (c *Client) Close() error {
	return c.rpc.Close()
}

// ChainID returns the current chain id.
func (c *Client) ChainID() (*big.Int, error) {
	return c.rpc.Eth().ChainID()
}

// BlockNumber returns the latest block number.
func (c *Client) BlockNumber() (uint64, error) {
	return c.rpc.Eth().BlockNumber()
}

// BlockByNumber fetches a full block by number (nil -> latest).
func (c *Client) BlockByNumber(n ethgo.BlockNumber, full bool) (*ethgo.Block, error) {
	return c.rpc.Eth().GetBlockByNumber(n, full)
}

// BalanceAt reads an account balance at a block (nil -> latest).
func (c *Client) BalanceAt(addr string, block ethgo.BlockNumberOrHash) (*big.Int, error) {
	return c.rpc.Eth().GetBalance(ethgo.HexToAddress(addr), block)
}

// NonceAt returns the account nonce at a given block (nil -> latest).
func (c *Client) NonceAt(addr string, block ethgo.BlockNumberOrHash) (uint64, error) {
	return c.rpc.Eth().GetNonce(ethgo.HexToAddress(addr), block)
}

/* ---------- Gas/fees ---------- */

// SuggestGasPrice returns the legacy gas price (pre-1559 fallback).
func (c *Client) SuggestGasPrice() (*big.Int, error) {
	maxGasPrice, err := c.rpc.Eth().GasPrice()
	if err != nil {
		return nil, err
	}
	return utils.U64ToBig(maxGasPrice), nil
}

// SuggestGasTipCap returns the EIP-1559 priority fee per gas.
func (c *Client) SuggestGasTipCap() (*big.Int, error) {
	var out string
	// Not all nodes support eth_maxPriorityFeePerGas, so handle fallback outside.
	if err := c.rpc.Call("eth_maxPriorityFeePerGas", &out); err != nil {
		return nil, err
	}
	return utils.StrToBig(out)
}

// FeeHistory returns EIP-1559 fee history.
func (c *Client) FeeHistory(from, to ethgo.BlockNumber) (*jsonrpc.FeeHistory, error) {
	return c.rpc.Eth().FeeHistory(from, to)
}

/* ---------- Call / Estimate ---------- */

// CallMsg mirrors the common EVM call message; maps directly to ethgo.CallMsg.
type CallMsg = ethgo.CallMsg

// EstimateGas simulates a tx and returns the needed gas.
func (c *Client) EstimateGas(msg *CallMsg) (uint64, error) {
	return c.rpc.Eth().EstimateGas(msg)
}

// Call executes a read-only call (eth_call).
func (c *Client) Call(msg *CallMsg, block ethgo.BlockNumber) (string, error) {
	return c.rpc.Eth().Call(msg, block)
}

/* ---------- Tx send / receipt ---------- */

// SendRawTransaction broadcasts a signed raw tx (RLP) and returns its hash.
func (c *Client) SendRawTransaction(rawTx []byte) (ethgo.Hash, error) {
	return c.rpc.Eth().SendRawTransaction(rawTx)
}

// TransactionReceipt fetches the receipt for a mined tx (may be nil before mined).
func (c *Client) TransactionReceipt(h ethgo.Hash) (*ethgo.Receipt, error) {
	return c.rpc.Eth().GetTransactionReceipt(h)
}

/* ---------- Logs / Subscribe (WS endpoint required) ---------- */

// FilterQuery is re-exported for convenience.
type FilterQuery = ethgo.LogFilter

// FilterLogs executes a one-off logs query (eth_getLogs).
func (c *Client) FilterLogs(q *FilterQuery) ([]*ethgo.Log, error) {
	return c.rpc.Eth().GetLogs(q)
}

/* ---------- Helpers ---------- */

// Address, Hash, and common types re-export (optional, ergonomic).
type (
	Address = ethgo.Address
	Hash    = ethgo.Hash
	Header  = ethgo.Block
	Block   = ethgo.Block
	Receipt = ethgo.Receipt
	Log     = ethgo.Log
)
