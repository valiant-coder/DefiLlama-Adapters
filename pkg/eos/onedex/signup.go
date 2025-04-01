package onedex

import (
	"context"

	"github.com/eoscanada/eos-go"
)

/*
{
name: applyacc,
base: ,
fields: [
{
name: id,
type: uint64
},
{
name: pubkey,
type: string
}
]
},
*/

// 调用eos合约的 signup.1dex 的appleyacc action

// ApplyAccArgs represents arguments for the applyacc action
type ApplyAccArgs struct {
	ID     uint64 `json:"id"`
	Pubkey string `json:"pubkey"`
}

// SignupClient encapsulates functionality for interacting with the signup.1dex contract
type SignupClient struct {
	api        *eos.API
	contract   string
	actor      string
	priv       string
	permission string
}

// NewSignupClient creates a new client for interacting with the signup.1dex contract
func NewSignupClient(endpoint string, contract string, actor string, priv string, permission string) *SignupClient {
	api := eos.New(endpoint)
	return &SignupClient{
		api:        api,
		actor:      actor,
		priv:       priv,
		permission: permission,
		contract:   contract,
	}
}

// ApplyAcc calls the applyacc action on the signup.1dex contract
func (c *SignupClient) ApplyAcc(ctx context.Context, uid uint64, pubkey string) (*eos.PushTransactionFullResp, error) {
	keyBag := &eos.KeyBag{}
	err := keyBag.ImportPrivateKey(ctx, c.priv)
	if err != nil {
		return nil, err
	}
	c.api.SetSigner(keyBag)

	action := &eos.Action{
		Account: eos.AN(c.contract),
		Name:    eos.ActN("applyacc"),
		Authorization: []eos.PermissionLevel{
			{Actor: eos.AN(c.actor), Permission: eos.PN(c.permission)},
		},
		ActionData: eos.NewActionData(ApplyAccArgs{
			ID:     uid,
			Pubkey: pubkey,
		}),
	}

	return c.api.SignPushActions(ctx, action)
}

type UIDPubkey struct {
	ID     uint64 `json:"id"`
	Pubkey string `json:"pubkey"`
	Status uint8  `json:"status"`
}

// 获取 accapplies表记录，scope是 sigup.1dex, index_position 是primary key_type 是i64,
func (c *SignupClient) GetPubkeyByUID(ctx context.Context, uid string) (string, error) {
	request := eos.GetTableRowsRequest{
		Code:       c.contract,
		Scope:      c.contract,
		Table:      "accapplies",
		Index:      "1",
		KeyType:    "i64",
		LowerBound: uid,
		UpperBound: uid,
		JSON:       true,
		Limit:      100,
	}

	response, err := c.api.GetTableRows(ctx, request)
	if err != nil {
		return "", err
	}

	var uidPubkeys []*UIDPubkey
	err = response.JSONToStructs(&uidPubkeys)
	if err != nil {
		return "", err
	}
	if len(uidPubkeys) == 0 {
		return "", nil
	}
	return uidPubkeys[0].Pubkey, nil
}
