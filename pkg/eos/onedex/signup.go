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
	actor      string
	priv       string
	permission string
}

// NewSignupClient creates a new client for interacting with the signup.1dex contract
func NewSignupClient(endpoint string, actor string, priv string, permission string) *SignupClient {
	api := eos.New(endpoint)
	return &SignupClient{
		api:        api,
		actor:      actor,
		priv:       priv,
		permission: permission,
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
		Account: eos.AN("signup.1dex"),
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
