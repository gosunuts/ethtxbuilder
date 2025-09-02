package transaction

import (
	"math/big"

	"github.com/gosunuts/ethtxbuilder/client"
	"github.com/gosunuts/ethtxbuilder/utils"
)

func BroadcastTx(client *client.Client, from string, to string, amount *big.Int, sign utils.SignFunc) (string, error) {
	nonce, err := client.NonceManager.Next(from)
	if err != nil {
		return "", err
	}
	maxPriorityFeePerGas, err := client.SuggestGasTipCap()
	if err != nil {
		return "", err
	}

	maxFeePerGas, err := client.SuggestGasPrice()
	if err != nil {
		return "", err
	}

	gasLimit := 21000

	rawTx, err := NewTransferTx(client.ChainId, nonce, to, amount, uint64(gasLimit), maxPriorityFeePerGas, maxFeePerGas, nil, sign)

	txhash, err := client.SendRawTransaction(rawTx)
	if err != nil {
		return "", err
	}
	return txhash.String(), nil
}

func NewTransferTx(chainId *big.Int, nonce uint64, to string, amount *big.Int, gasLimit uint64, maxPriorityFeePerGas *big.Int, maxFeePerGas *big.Int, data []byte, sign utils.SignFunc) ([]byte, error) {
	tx := NewDynamicTx(chainId, nonce, to, amount, gasLimit, maxFeePerGas, maxFeePerGas, nil)
	err := tx.Sign(sign)
	if err != nil {
		return nil, err
	}
	return tx.EncodeRLP(), nil
}
