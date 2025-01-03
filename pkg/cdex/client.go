package dex

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/eoscanada/eos-go"
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
	Symbol string `json:"sym"`
	Contract string `json:"contract"`
}

type Pool struct {
	ID             uint64            `json:"id"`
	Base           EOSExtendedAsset  `json:"base"`
	Quote          EOSExtendedAsset  `json:"quote"`
	AskingTime     string            `json:"asking_time"`
	TradingTime    string            `json:"trading_time"`
	MinAmount      uint64            `json:"min_amount"`
	MaxFlct        uint64            `json:"max_flct"`
	PricePrecision uint8             `json:"price_precision"`
	TakerFeeRate   string            `json:"taker_fee_rate"`
	MakerFeeRate   string            `json:"maker_fee_rate"`
	Status         uint8             `json:"status"`
}


type Order struct {
	ID       uint64    `json:"id"`
	App      eos.Name  `json:"app"`
	CID      string    `json:"cid"`
	Trader   struct {
		Actor eos.AccountName `json:"actor"`
		Permission eos.PermissionName `json:"permission"`
	} `json:"trader"`
	Price    uint64    `json:"price"`
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

func (c *Client) GetOrders(ctx context.Context, poolID uint64, isBid bool) ([]Order, error) {
	tableName := "asks"
	if isBid {
		tableName = "bids"
	}

	request := eos.GetTableRowsRequest{
		Code:       c.dexContract,
		Scope:      fmt.Sprintf("%d", poolID),
		Table:      tableName,
		JSON:       true,
		LowerBound: "0",
		UpperBound: "-1",
		Limit:      1000,
	}

	response, err := c.api.GetTableRows(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("failed to get orders: %w", err)
	}

	var orders []Order
	err = json.Unmarshal(response.Rows, &orders)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal orders: %w", err)
	}

	return orders, nil
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
