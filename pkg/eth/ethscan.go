package eth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/shopspring/decimal"
	"github.com/spf13/cast"
	"golang.org/x/sync/errgroup"
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
	type result struct {
		tokenBalances []TokenBalance
		nativeBalance string
		err           error
	}

	ctxWithTimeout, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	g, ctx := errgroup.WithContext(ctxWithTimeout)

	tokenCh := make(chan result, 1)
	nativeCh := make(chan result, 1)

	g.Go(func() error {
		url := fmt.Sprintf("%s/api/v2/addresses/%s/token-balances", c.Endpoint, address)
		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			tokenCh <- result{err: err}
			return err
		}
		req.Header.Set("x-api-key", c.ApiKey)

		var resp *http.Response
		for retries := 0; retries < 3; retries++ {
			resp, err = http.DefaultClient.Do(req)
			if err == nil {
				break
			}
			time.Sleep(time.Duration(retries+1) * 100 * time.Millisecond)
		}
		if err != nil {
			tokenCh <- result{err: fmt.Errorf("failed to get token balances after retries: %v", err)}
			return err
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			tokenCh <- result{err: err}
			return err
		}

		var response []TokenBalanceResponse
		if err = json.Unmarshal(body, &response); err != nil {
			tokenCh <- result{err: err}
			return err
		}

		tokenBalances := make([]TokenBalance, 0, len(response))
		for _, balance := range response {
			decimals := cast.ToInt32(balance.Token.Decimals)
			tokenBalances = append(tokenBalances, TokenBalance{
				TokenAddress:  balance.Token.Address,
				TokenSymbol:   balance.Token.Symbol,
				TokenName:     balance.Token.Name,
				TokenDecimals: decimals,
				Balance:       decimal.RequireFromString(balance.Value).Shift(-decimals),
				Type:          balance.Token.Type,
			})
		}
		tokenCh <- result{tokenBalances: tokenBalances}
		return nil
	})

	g.Go(func() error {
		nativeBalance, err := c.GetNativeTokenBalanceByAddress(ctx, address)
		if err != nil {
			nativeCh <- result{err: err}
			return err
		}
		nativeCh <- result{nativeBalance: nativeBalance}
		return nil
	})

	if err := g.Wait(); err != nil {
		return nil, err
	}

	tokenResult := <-tokenCh
	if tokenResult.err != nil {
		return nil, tokenResult.err
	}

	nativeResult := <-nativeCh
	if nativeResult.err != nil {
		return nil, nativeResult.err
	}

	tokenResult.tokenBalances = append(tokenResult.tokenBalances, TokenBalance{
		TokenAddress:  "0x0000000000000000000000000000000000000000",
		TokenSymbol:   "BTC",
		TokenName:     "Bitcoin",
		TokenDecimals: 18,
		Balance:       decimal.RequireFromString(nativeResult.nativeBalance).Shift(-18),
		Type:          "native",
	})

	return tokenResult.tokenBalances, nil
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
