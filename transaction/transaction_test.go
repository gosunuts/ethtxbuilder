package transaction

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/gosunuts/ethtxbuilder/client"
	"github.com/gosunuts/ethtxbuilder/utils"
	"github.com/stretchr/testify/require"
)

func TestBroadcastTx(t *testing.T) {
	client, err := client.NewClient(client.EthereumSepolia)
	require.NoError(t, err)

	from := ""
	to := ""
	signfunc := utils.NewRawPrivateSigner("")

	txid, err := BroadcastTx(client, from, to, big.NewInt(1e15), signfunc)
	require.NoError(t, err)
	fmt.Println(txid)
}
