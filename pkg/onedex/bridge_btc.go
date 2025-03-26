package onedex

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/eoscanada/eos-go"
)

type BTCMappingAddrRequest struct {
	Remark           string `json:"remark"`
	RecipientAddress string `json:"recipient_address"`
}

type BTCBridgeClient struct {
	api    *eos.API
	bridge string
	actor  string
	priv   string
}

func NewBTCBridgeClient(endpoint string, bridge string, actor string, priv string) *BTCBridgeClient {
	api := eos.New(endpoint)
	return &BTCBridgeClient{
		api:    api,
		bridge: bridge,
		actor:  actor,
		priv:   priv,
	}
}

func (c *BTCBridgeClient) MappingAddress(ctx context.Context, req BTCMappingAddrRequest) (*eos.PushTransactionFullResp, error) {
	keyBag := &eos.KeyBag{}
	err := keyBag.ImportPrivateKey(ctx, c.priv)
	if err != nil {
		return nil, err
	}
	c.api.SetSigner(keyBag)

	action := &eos.Action{
		Account: eos.AN(c.bridge),
		Name:    eos.ActN("appaddrmap"),
		Authorization: []eos.PermissionLevel{
			{Actor: eos.AN(c.actor), Permission: eos.PN("active")},
		},
		ActionData: eos.NewActionData(struct {
			Actor                eos.AccountName `eos:"actor"`
			PermissionID         uint64          `eos:"permission_id"`
			RecipientAddress     string          `eos:"recipient_address"`
			Remark               string          `eos:"remark"`
			AssignDepositAddress string          `eos:"assign_deposit_address"`
		}{
			Actor:                eos.AN(c.actor),
			PermissionID:         0,
			Remark:               req.Remark,
			RecipientAddress:     req.RecipientAddress,
			AssignDepositAddress: "",
		}),
	}

	return c.api.SignPushActions(ctx, action)
}

type RequestBTCDepositAddress struct {
	Remark              string `json:"remark"`
	RecipientEVMAddress string `json:"recipient_evm_address"`
}

func makeBTCKey256(recipientEVMAddress, remark string) [32]byte {
	recipientEVMAddress = strings.TrimPrefix(recipientEVMAddress, "0x")
	key := recipientEVMAddress + "-" + remark
	key = strings.ToLower(key)
	return sha256.Sum256([]byte(key))
}

func (c *BTCBridgeClient) GetDepositAddress(ctx context.Context, req RequestBTCDepositAddress) (string, error) {
	key := makeBTCKey256(req.RecipientEVMAddress, req.Remark)

	request := eos.GetTableRowsRequest{
		Code:       c.bridge,
		Scope:      "0",
		Table:      "addrmappings",
		LowerBound: hex.EncodeToString(key[:]),
		UpperBound: hex.EncodeToString(key[:]),
		Index:      "6",
		KeyType:    "sha256",
		JSON:       true,
		Limit:      1,
	}

	response, err := c.api.GetTableRows(ctx, request)
	if err != nil {
		return "", fmt.Errorf("get table rows failed: %w", err)
	}

	var rows []struct {
		ID      uint64 `json:"id"`
		BTCAddr string `json:"btc_address"`
	}
	err = json.Unmarshal(response.Rows, &rows)
	if err != nil {
		return "", fmt.Errorf("unmarshal rows failed: %w", err)
	}

	if len(rows) == 0 {
		return "", fmt.Errorf("deposit address not found")
	}

	return rows[0].BTCAddr, nil
}
