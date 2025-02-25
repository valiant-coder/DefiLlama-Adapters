package eth

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/shopspring/decimal"
)

type Client struct {
	ethClient *ethclient.Client
}

func NewClient(url string) (*Client, error) {
	client, err := ethclient.Dial(url)
	if err != nil {
		return nil, err
	}
	return &Client{ethClient: client}, nil
}

func (c *Client) TransferETH(fromAddr string, toAddr string, amount float64, privateKey string) (string, error) {
	fromAddress := common.HexToAddress(fromAddr)
	toAddress := common.HexToAddress(toAddr)

	nonce, err := c.ethClient.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		return "", err
	}

	gasPrice, err := c.ethClient.SuggestGasPrice(context.Background())
	if err != nil {
		return "", err
	}
	gasPriceDecimal := decimal.NewFromBigInt(gasPrice, 0).Mul(decimal.NewFromFloat(1.25))
	gasPrice = gasPriceDecimal.BigInt()


	chainID, err := c.ethClient.ChainID(context.Background())
	if err != nil {
		return "", err
	}

	amountBigInt := decimal.NewFromFloat(amount).Shift(18).BigInt()

	tx := types.NewTransaction(nonce, toAddress, amountBigInt, 42000, gasPrice, nil)

	privateKeyECDSA, err := crypto.HexToECDSA(privateKey)
	if err != nil {
		return "", err
	}

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKeyECDSA)
	if err != nil {
		return "", err
	}

	err = c.ethClient.SendTransaction(context.Background(), signedTx)
	if err != nil {
		return "", err
	}

	return signedTx.Hash().Hex(), nil
}
