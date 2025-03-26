package onedex

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/eoscanada/eos-go"
	"github.com/spf13/cast"
)

type MappingAddrRequest struct {
	PermissionID     uint64 `json:"permission_id"`
	Remark           string `json:"remark"`
	RecipientAddress string `json:"recipient_address"`
}

type BridgeClient struct {
	api    *eos.API
	bridge string
	actor  string
	priv   string
}

func NewBridgeClient(endpoint string, bridge string, actor string, priv string) *BridgeClient {
	api := eos.New(endpoint)
	return &BridgeClient{
		api:    api,
		bridge: bridge,
		actor:  actor,
		priv:   priv,
	}
}

func (c *BridgeClient) MappingAddress(ctx context.Context, req MappingAddrRequest) (*eos.PushTransactionFullResp, error) {
	keyBag := &eos.KeyBag{}
	err := keyBag.ImportPrivateKey(ctx, c.priv)
	if err != nil {
		return nil, err
	}
	c.api.SetSigner(keyBag)

	action := &eos.Action{
		Account: eos.AN(c.bridge),
		Name:    eos.ActN("mappingaddr"),
		Authorization: []eos.PermissionLevel{
			{Actor: eos.AN(c.actor), Permission: eos.PN("active")},
		},
		ActionData: eos.NewActionData(struct {
			Actor                eos.AccountName `eos:"actor"`
			PermissionID         uint64          `eos:"permission_id"`
			RecipientAddress     string          `eos:"recipient_address"`
			Remark               string          `eos:"remark"`
			WhitelistAddress     string          `eos:"whitelist_address"`
			AssignDepositAddress string          `eos:"assign_deposit_address"`
		}{
			Actor:                eos.AN(c.actor),
			PermissionID:         req.PermissionID,
			Remark:               req.Remark,
			RecipientAddress:     req.RecipientAddress,
			WhitelistAddress:     "",
			AssignDepositAddress: "",
		}),
	}

	return c.api.SignPushActions(ctx, action)
}

type RequestDepositAddress struct {
	PermissionID uint64 `json:"permission_id"`
	Remark       string `json:"remark"`
	Recipient    string `json:"recipient"`
}

func makeKey256(recipientAddress, remark string) [32]byte {
	key := recipientAddress + "-" + remark
	key = strings.ToLower(key)
	return sha256.Sum256([]byte(key))
}

func (c *BridgeClient) GetDepositAddress(ctx context.Context, req RequestDepositAddress) (string, error) {
	key := makeKey256(req.Recipient, req.Remark)

	request := eos.GetTableRowsRequest{
		Code:       c.bridge,
		Scope:      cast.ToString(req.PermissionID),
		Table:      "addrmappings",
		LowerBound: hex.EncodeToString(key[:]),
		UpperBound: hex.EncodeToString(key[:]),
		Index:      "4",
		KeyType:    "sha256",
		JSON:       true,
		Limit:      1,
	}

	response, err := c.api.GetTableRows(ctx, request)
	if err != nil {
		return "", fmt.Errorf("get table rows failed: %w", err)
	}

	var rows []struct {
		ID             uint64 `json:"id"`
		DepositAddress string `json:"deposit_address"`
	}
	err = json.Unmarshal(response.Rows, &rows)
	if err != nil {
		return "", fmt.Errorf("unmarshal rows failed: %w", err)
	}

	if len(rows) == 0 {
		return "", fmt.Errorf("deposit address not found")
	}

	return rows[0].DepositAddress, nil
}
