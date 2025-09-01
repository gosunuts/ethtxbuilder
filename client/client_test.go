package client

import (
	"math/big"
	"testing"

	"github.com/gosunuts/ethtxbuilder/utils"
	"github.com/stretchr/testify/require"
	"github.com/umbracle/ethgo"
)

const (
	EthereumMainnet = "https://ethereum-rpc.publicnode.com"
	EthereumSepolia = "https://ethereum-sepolia-rpc.publicnode.com"
	OptimismMainnet = "https://optimism-rpc.publicnode.com"
	OptimismSepolia = "https://optimism-sepolia-rpc.publicnode.com"
	ArbitrumOne     = "https://arbitrum-one-rpc.publicnode.com"
	ArbitrumSepolia = "https://arbitrum-sepolia-rpc.publicnode.com"
)

func TestClientEthMainnet(t *testing.T) {
	for _, test := range []struct {
		name     string
		endpoint string
		chainID  *big.Int
	}{
		{"EthereumMainnet", EthereumMainnet, big.NewInt(1)},
		{"EthereumSepolia", EthereumSepolia, big.NewInt(11155111)},
		{"OptimismMainnet", OptimismMainnet, big.NewInt(10)},
		{"OptimismSepolia", OptimismSepolia, big.NewInt(11155420)},
		{"ArbitrumOne", ArbitrumOne, big.NewInt(42161)},
		{"ArbitrumSepolia", ArbitrumSepolia, big.NewInt(421614)},
	} {
		cli, err := New(test.endpoint)
		require.NoError(t, err, "New client", test.name)
		defer cli.Close()

		// ChainID
		cid, err := cli.ChainID()
		require.NoError(t, err, "ChainID", test.name)
		require.Equal(t, 0, cid.Cmp(test.chainID), "unexpected chain id", test.name)

		// BlockNumber
		bn, err := cli.BlockNumber()
		require.NoError(t, err, "BlockNumber", test.name)
		require.Greater(t, bn, uint64(0), "unexpected block number", test.name)

		// Balance & Nonce at latest for a well-formed EOA (doesn't need funds)
		addr := utils.StrToAddr("0x0000000000000000000000000000000000000001")
		bal, err := cli.BalanceAt(addr, ethgo.Latest)
		require.NoError(t, err, "BalanceAt", test.name)
		require.NotNil(t, bal)

		_, err = cli.NonceAt(addr, ethgo.Latest)
		require.NoError(t, err, "NonceAt", test.name)

		gp, err := cli.SuggestGasPrice()
		require.NoError(t, err, "SuggestGasPrice", test.name)
		require.Greater(t, gp, uint64(0), "unexpected gas price", test.name)

		// tip, err := cli.SuggestGasTipCap()
		// require.NoError(t, err, "SuggestGasTipCap", test.name)
		// require.Greater(t, tip, uint64(0), "unexpected gas tip cap", test.name)
	}
}
