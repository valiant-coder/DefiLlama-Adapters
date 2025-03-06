package eth

import (
	"context"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/shopspring/decimal"
)

const erc20ABI = `[
	{
		"constant": true,
		"inputs": [],
		"name": "decimals",
		"outputs": [
			{
				"name": "",
				"type": "uint8"
			}
		],
		"payable": false,
		"stateMutability": "view",
		"type": "function"
	},
	{
		"constant": false,
		"inputs": [
			{
				"name": "_to",
				"type": "address"
			},
			{
				"name": "_value",
				"type": "uint256"
			}
		],
		"name": "transfer",
		"outputs": [
			{
				"name": "",
				"type": "bool"
			}
		],
		"payable": false,
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"constant": false,
		"inputs": [
			{
				"name": "_to",
				"type": "address"
			},
			{
				"name": "_value",
				"type": "uint256"
			}
		],
		"name": "mint",
		"outputs": [
			{
				"name": "",
				"type": "bool"
			}
		],
		"payable": false,
		"stateMutability": "nonpayable",
		"type": "function"
	}
]`

type Token struct {
	Address common.Address
	ABI     abi.ABI
}

func NewToken(addr string) (*Token, error) {
	address := common.HexToAddress(addr)
	parsedABI, err := abi.JSON(strings.NewReader(erc20ABI))
	if err != nil {
		return nil, err
	}
	return &Token{Address: address, ABI: parsedABI}, nil
}

func (t *Token) GetDecimals(c *Client) (uint8, error) {
	data, err := t.ABI.Pack("decimals")
	if err != nil {
		return 0, err
	}

	msg := ethereum.CallMsg{
		To:   &t.Address,
		Data: data,
	}
	result, err := c.ethClient.CallContract(context.Background(), msg, nil)
	if err != nil {
		return 0, err
	}

	var decimals uint8
	if err := t.ABI.UnpackIntoInterface(&decimals, "decimals", result); err != nil {
		return 0, err
	}

	return decimals, nil
}

func (t *Token) TransferERC20(c *Client, fromAddr string, toAddr string, amount float64, privateKey string) (string, error) {
	fromAddress := common.HexToAddress(fromAddr)
	toAddress := common.HexToAddress(toAddr)
	decimals, err := t.GetDecimals(c)
	if err != nil {
		return "", err
	}

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

	amountBigInt := decimal.NewFromFloat(amount).Shift(int32(decimals)).BigInt()

	data, err := t.ABI.Pack("transfer", toAddress, amountBigInt)
	if err != nil {
		return "", err
	}

	tx := types.NewTransaction(nonce, t.Address, big.NewInt(0), 100000, gasPrice, data)

	privateKeyECDSA, err := crypto.HexToECDSA(privateKey)
	if err != nil {
		return "", err
	}

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKeyECDSA)
	if err != nil {
		return "", err
	}

	if err := c.ethClient.SendTransaction(context.Background(), signedTx); err != nil {
		return "", err
	}

	return signedTx.Hash().Hex(), nil
}

func (t *Token) MintERC20(c *Client,
	ownerAddr string,
	toAddr string,
	amount float64,
	privateKey string,
	nonce uint64,
) (string, error) {
	ownerAddress := common.HexToAddress(ownerAddr)
	toAddress := common.HexToAddress(toAddr)
	decimals, err := t.GetDecimals(c)
	if err != nil {
		return "", err
	}

	if nonce == 0 {
		nonce, err = c.ethClient.PendingNonceAt(context.Background(), ownerAddress)
		if err != nil {
			return "", err
		}
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

	amountBigInt := decimal.NewFromFloat(amount).Shift(int32(decimals)).BigInt()

	data, err := t.ABI.Pack("mint", toAddress, amountBigInt)
	if err != nil {
		return "", err
	}

	tx := types.NewTransaction(nonce, t.Address, big.NewInt(0), 100000, gasPrice, data)

	privateKeyECDSA, err := crypto.HexToECDSA(privateKey)
	if err != nil {
		return "", err
	}

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKeyECDSA)
	if err != nil {
		return "", err
	}

	if err := c.ethClient.SendTransaction(context.Background(), signedTx); err != nil {
		return "", err
	}

	return signedTx.Hash().Hex(), nil
}
