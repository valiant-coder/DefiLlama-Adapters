package eth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/shopspring/decimal"
	"github.com/spf13/cast"
)

type EthScanClient struct {
	ApiKey   string
	Endpoint string
}

func NewEthScanClient(endpoint, apiKey string) *EthScanClient {
	return &EthScanClient{
		ApiKey:   apiKey,
		Endpoint: endpoint,
	}
}

type TokenBalance struct {
	TokenAddress  string          `json:"token_address"`
	TokenSymbol   string          `json:"token_symbol"`
	TokenName     string          `json:"token_name"`
	TokenDecimals int32           `json:"token_decimals"`
	Balance       decimal.Decimal `json:"balance"`
	Type          string          `json:"type"`
}

/*
https://scan-testnet.exsat.network/api/v2/addresses/0x2614e5588275b02B23CdbeFed8E5dA6D2f59d1c6/token-balances

[

	{
	  "token": {
	    "address": "0x4aa4365da82ACD46e378A6f3c92a863f3e763d34",
	    "circulating_market_cap": null,
	    "decimals": "18",
	    "exchange_rate": null,
	    "holders": "1308",
	    "icon_url": null,
	    "name": "Wrapped BTC",
	    "symbol": "XBTC",
	    "total_supply": "2139706762002812912765",
	    "type": "ERC-20",
	    "volume_24h": null
	  },
	  "token_id": null,
	  "token_instance": null,
	  "value": "1000000000000000000"
	}

]
*/
type TokenBalanceResponse struct {
	Token struct {
		Address  string `json:"address"`
		Symbol   string `json:"symbol"`
		Name     string `json:"name"`
		Decimals string `json:"decimals"`
		Type     string `json:"type"`
	} `json:"token"`
	Value string `json:"value"`
}

func (c *EthScanClient) GetTokenBalancesByAddress(ctx context.Context, address string) ([]TokenBalance, error) {

	url := fmt.Sprintf("%s/api/v2/addresses/%s/token-balances", c.Endpoint, address)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("x-api-key", c.ApiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var response []TokenBalanceResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}

	tokenBalances := make([]TokenBalance, len(response))
	for i, balance := range response {
		tokenBalances[i] = TokenBalance{
			TokenAddress:  balance.Token.Address,
			TokenSymbol:   balance.Token.Symbol,
			TokenName:     balance.Token.Name,
			TokenDecimals: cast.ToInt32(balance.Token.Decimals),
			Balance:       decimal.RequireFromString(balance.Value).Shift(-cast.ToInt32(balance.Token.Decimals)),
			Type:          balance.Token.Type,
		}
	}

	nativeBalance, err := c.GetNativeTokenBalanceByAddress(ctx, address)
	if err != nil {
		return nil, err
	}

	tokenBalances = append(tokenBalances, TokenBalance{
		TokenAddress:  "0x0000000000000000000000000000000000000000",
		TokenSymbol:   "BTC",
		TokenName:     "Bitcoin",
		TokenDecimals: 18,
		Balance:       decimal.RequireFromString(nativeBalance).Shift(-18),
		Type:          "native",
	})

	return tokenBalances, nil
}

/*
https://scan2.exactsat.io/api/v2/addresses/xxx
*/

type NativeTokenBalance struct {
	CoinBalance string `json:"coin_balance"`
}

func (c *EthScanClient) GetNativeTokenBalanceByAddress(ctx context.Context, address string) (string, error) {
	url := fmt.Sprintf("%s/api/v2/addresses/%s", c.Endpoint, address)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("x-api-key", c.ApiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var response NativeTokenBalance
	err = json.Unmarshal(body, &response)
	if err != nil {
		return "", err
	}

	return response.CoinBalance, nil
}
