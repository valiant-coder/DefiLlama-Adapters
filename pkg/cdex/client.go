package cdex

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/eoscanada/eos-go"
	"github.com/spf13/cast"
)

type Client struct {
	api          *eos.API
	dexContract  string
	poolContract string
}

func NewClient(nodeUrl string, dexContract string, poolContract string) *Client {
	api := eos.New(nodeUrl)
	return &Client{
		api:          api,
		dexContract:  dexContract,
		poolContract: poolContract,
	}
}

type EOSExtendedAsset struct {
	Symbol   string `json:"sym"`
	Contract string `json:"contract"`
}

func (e *EOSExtendedAsset) SymbolAndPrecision() (string, uint8) {
	parts := strings.Split(e.Symbol, ",")
	if len(parts) != 2 {
		return "", 0
	}
	precision := cast.ToUint8(parts[0])
	symbol := parts[1]
	return symbol, precision
}

type Pool struct {
	ID             uint64           `json:"id"`
	Base           EOSExtendedAsset `json:"base"`
	Quote          EOSExtendedAsset `json:"quote"`
	AskingTime     string           `json:"asking_time"`
	TradingTime    string           `json:"trading_time"`
	MinAmount      uint64           `json:"min_amount"`
	MaxFlct        uint64           `json:"max_flct"`
	PricePrecision uint8            `json:"price_precision"`
	TakerFeeRate   string           `json:"taker_fee_rate"`
	MakerFeeRate   string           `json:"maker_fee_rate"`
	Status         uint8            `json:"status"`
}

type Order struct {
	ID     uint64   `json:"id"`
	App    eos.Name `json:"app"`
	CID    string   `json:"cid"`
	Trader struct {
		Actor      eos.AccountName    `json:"actor"`
		Permission eos.PermissionName `json:"permission"`
	} `json:"trader"`
	Price    string    `json:"price"`
	Quantity eos.Asset `json:"quantity"`
	Filled   eos.Asset `json:"filled"`
	IsBid    uint8     `json:"is_bid"`
}

type Fund struct {
	PoolID uint64    `json:"pool_id"`
	Base   eos.Asset `json:"base"`
	Quote  eos.Asset `json:"quote"`
}

func (c *Client) GetPools(ctx context.Context) ([]Pool, error) {
	request := eos.GetTableRowsRequest{
		Code:       c.poolContract,
		Scope:      c.poolContract,
		Table:      "pools",
		JSON:       true,
		LowerBound: "0",
		UpperBound: "-1",
		Limit:      1000,
	}

	response, err := c.api.GetTableRows(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("failed to get pools: %w", err)
	}

	var pools []Pool
	err = json.Unmarshal(response.Rows, &pools)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal pools: %w", err)
	}

	return pools, nil
}

func (c *Client) getOrdersPage(ctx context.Context, poolID uint64, isBid bool, lowerBound string) ([]Order, string, error) {
	tableName := "asks"
	if isBid {
		tableName = "bids"
	}

	request := eos.GetTableRowsRequest{
		Code:       c.dexContract,
		Scope:      fmt.Sprintf("%d", poolID),
		Table:      tableName,
		JSON:       true,
		LowerBound: lowerBound,
		UpperBound: "-1",
		Limit:      500,
	}

	response, err := c.api.GetTableRows(ctx, request)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get orders: %w", err)
	}

	var orders []Order
	err = json.Unmarshal(response.Rows, &orders)
	if err != nil {
		return nil, "", fmt.Errorf("failed to unmarshal orders: %w", err)
	}

	nextLowerBound := ""
	if response.More && len(orders) > 0 {
		nextLowerBound = fmt.Sprintf("%d", orders[len(orders)-1].ID+1)
	}

	return orders, nextLowerBound, nil
}

func (c *Client) GetOrders(ctx context.Context, poolID uint64, isBid bool) ([]Order, error) {
	var allOrders []Order
	lowerBound := "0"

	for {
		orders, nextLowerBound, err := c.getOrdersPage(ctx, poolID, isBid, lowerBound)
		if err != nil {
			return nil, err
		}

		allOrders = append(allOrders, orders...)

		if nextLowerBound == "" {
			break
		}
		lowerBound = nextLowerBound
	}

	return allOrders, nil
}

func (c *Client) GetUserFunds(ctx context.Context, account string) ([]Fund, error) {
	request := eos.GetTableRowsRequest{
		Code:       c.dexContract,
		Scope:      account,
		Table:      "funds",
		JSON:       true,
		LowerBound: "0",
		UpperBound: "-1",
		Limit:      1000,
	}

	response, err := c.api.GetTableRows(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("failed to get funds: %w", err)
	}
	fmt.Println(string(response.Rows))

	var funds []Fund
	err = json.Unmarshal(response.Rows, &funds)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal funds: %w", err)
	}

	return funds, nil
}
