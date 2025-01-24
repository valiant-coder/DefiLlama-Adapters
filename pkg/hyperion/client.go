package hyperion

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/shopspring/decimal"
)

type Client struct {
	endpoint   string
	httpClient *http.Client
}

type GetActionsRequest struct {
	// Notified account
	Account string `json:"account"`
	// Filter by code:name
	Filter string `json:"filter,omitempty"`
	Track  int    `json:"track,omitempty"`
	Skip   int    `json:"skip,omitempty"`
	Limit  int    `json:"limit,omitempty"`
	// Sort direction ('desc', 'asc', '1' or '-1')
	Sort string `json:"sort,omitempty"`
	// Filter actions by block_num range [from]-[to]
	BlockNum       string `json:"block_num,omitempty"`
	GlobalSequence string `json:"global_sequence,omitempty"`
	After          string `json:"after,omitempty"`
	Before         string `json:"before,omitempty"`
	Simple         bool   `json:"simple,omitempty"`
	NoBinary       bool   `json:"noBinary,omitempty"`
	CheckLib       bool   `json:"checkLib,omitempty"`
}

type Action struct {
	Timestamp            string            `json:"@timestamp"`
	BlockNum             uint64            `json:"block_num"`
	TrxID                string            `json:"trx_id"`
	Act                  ActionData        `json:"act"`
	Receipts             []json.RawMessage `json:"receipts"`
	CpuUsageUs           int               `json:"cpu_usage_us"`
	AccountRamDeltas     []json.RawMessage `json:"account_ram_deltas"`
	GlobalSequence       uint64            `json:"global_sequence"`
	Producer             string            `json:"producer"`
	ActionOrdinal        int               `json:"action_ordinal"`
	CreatorActionOrdinal int               `json:"creator_action_ordinal"`
}

type Authorization struct {
	Actor      string `json:"actor"`
	Permission string `json:"permission"`
}
type ActionData struct {
	Account       string          `json:"account"`
	Name          string          `json:"name"`
	Authorization []Authorization `json:"authorization"`
	Data          json.RawMessage `json:"data"`
}

type SimpleAction struct {
	Block         uint64          `json:"block"`
	Timestamp     string          `json:"timestamp"`
	Contract      string          `json:"contract"`
	Action        string          `json:"action"`
	Actors        string          `json:"actors"`
	Notified      string          `json:"notified"`
	TransactionID string          `json:"transaction_id"`
	Data          json.RawMessage `json:"data"`
}

type GetActionsResponse struct {
	QueryTimeMs          float64 `json:"query_time_ms"`
	Cached               bool    `json:"cached"`
	Lib                  int     `json:"lib"`
	LastIndexedBlock     uint64  `json:"last_indexed_block"`
	LastIndexedBlockTime string  `json:"last_indexed_block_time"`
	Total                struct {
		Value    int    `json:"value"`
		Relation string `json:"relation"`
	} `json:"total"`
	SimpleActions []SimpleAction `json:"simple_actions"`
	Actions       []Action       `json:"actions"`
}

type GetDeltasResponse struct {
	QueryTimeMs          float64 `json:"query_time_ms"`
	LastIndexedBlock     uint64  `json:"last_indexed_block"`
	LastIndexedBlockTime string  `json:"last_indexed_block_time"`
	Total                struct {
		Value    int    `json:"value"`
		Relation string `json:"relation"`
	} `json:"total"`
	Deltas []Delta `json:"deltas"`
}

type Delta struct {
	Timestamp  string          `json:"timestamp"`
	Present    int             `json:"present"`
	Code       string          `json:"code"`
	Scope      string          `json:"scope"`
	Table      string          `json:"table"`
	PrimaryKey string          `json:"primary_key"`
	Payer      string          `json:"payer"`
	BlockNum   uint64          `json:"block_num"`
	BlockID    string          `json:"block_id"`
	Data       json.RawMessage `json:"data"`
}

func NewClient(endpoint string) *Client {
	return &Client{
		endpoint: endpoint,
		httpClient: &http.Client{
			Timeout: time.Second * 10,
		},
	}
}

func (c *Client) GetActions(ctx context.Context, req GetActionsRequest) (*GetActionsResponse, error) {
	baseUrl := fmt.Sprintf("%s/v2/history/get_actions", c.endpoint)

	params := url.Values{}

	if req.Account != "" {
		params.Add("account", req.Account)
	}
	if req.Filter != "" {
		params.Add("filter", req.Filter)
	}
	if req.Track != 0 {
		params.Add("track", strconv.Itoa(req.Track))
	}
	if req.Skip != 0 {
		params.Add("skip", strconv.Itoa(req.Skip))
	}
	if req.Limit != 0 {
		params.Add("limit", strconv.Itoa(req.Limit))
	}
	if req.Sort != "" {
		params.Add("sort", req.Sort)
	}
	if req.BlockNum != "" {
		params.Add("block_num", req.BlockNum)
	}
	if req.GlobalSequence != "" {
		params.Add("global_sequence", req.GlobalSequence)
	}
	if req.After != "" {
		params.Add("after", req.After)
	}
	if req.Before != "" {
		params.Add("before", req.Before)
	}
	if req.Simple {
		params.Add("simple", "true")
	}
	if req.NoBinary {
		params.Add("noBinary", "true")
	}
	if req.CheckLib {
		params.Add("checkLib", "true")
	}

	url := fmt.Sprintf("%s?%s", baseUrl, params.Encode())

	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request failed: %w", err)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("do request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var result GetActionsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response failed: %w", err)
	}

	return &result, nil
}

type GetTokensResponse struct {
	Tokens []Token `json:"tokens"`
}

type Token struct {
	Symbol    string          `json:"symbol"`
	Precision uint8           `json:"precision"`
	Amount    decimal.Decimal `json:"amount"`
	Contract  string          `json:"contract"`
}

func (c *Client) GetTokens(ctx context.Context, account string) ([]Token, error) {
	url := fmt.Sprintf("%s/v2/state/get_tokens?account=%s", c.endpoint, account)

	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request failed: %w", err)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("do request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var result GetTokensResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response failed: %w", err)
	}
	return result.Tokens, nil
}

type Transaction struct {
	TrxID   string   `json:"trx_id"`
	Actions []Action `json:"actions"`
}

func (c *Client) GetTransaction(ctx context.Context, txID string) (*Transaction, error) {
	url := fmt.Sprintf("%s/v2/history/get_transaction?id=%s", c.endpoint, txID)

	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request failed: %w", err)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("do request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var result Transaction
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response failed: %w", err)
	}
	return &result, nil
}

type EvmTxEvent struct {
	Event []interface{} `json:"event"`
}

type EvmTxData struct {
	EosEvmVersion string `json:"eos_evm_version"`
	RlpTx         string `json:"rlptx"`
	BaseFeePerGas string `json:"base_fee_per_gas"`
}

func (c *Client) GetEvmTxIDByEosTxID(eosTxID string) (string, error) {
	ctx := context.Background()
	tx, err := c.GetTransaction(ctx, eosTxID)
	if err != nil {
		return "", err
	}

	for _, action := range tx.Actions {
		if action.Act.Name == "evmtx" {
			var evmEvent EvmTxEvent
			if err := json.Unmarshal(action.Act.Data, &evmEvent); err != nil {
				return "", err
			}

			if len(evmEvent.Event) != 2 {
				return "", errors.New("invalid evmtx event format")
			}

			eventDataJSON, err := json.Marshal(evmEvent.Event[1])
			if err != nil {
				return "", err
			}

			var eventData EvmTxData
			if err := json.Unmarshal(eventDataJSON, &eventData); err != nil {
				return "", err
			}

			rlpTxBytes, err := hex.DecodeString(eventData.RlpTx)
			if err != nil {
				return "", err
			}

			hash := crypto.Keccak256(rlpTxBytes)

			return "0x" + hex.EncodeToString(hash), nil
		}
	}
	return "", nil
}
