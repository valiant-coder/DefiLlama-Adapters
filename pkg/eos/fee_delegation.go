package eos

import (
	"context"

	"github.com/eoscanada/eos-go"
	"github.com/eoscanada/eos-go/ecc"
	"github.com/eoscanada/eos-go/token"
)

// Version that separates user signature and payer signature
func CreateUserSignedTransaction(
	ctx context.Context,
	api *eos.API,
	fromAccount string,
	toAccount string,
	quantity string,
	memo string,
	userPrivateKeyWIF string,
	payerAccount string,
) (*eos.SignedTransaction, error) {
	asset, err := eos.NewEOSAssetFromString(quantity)
	if err != nil {
		return nil, err
	}

	// 1. Set up action
	action := &eos.Action{
		Account: eos.AN("eosio.token"),
		Name:    eos.ActN("transfer"),
		Authorization: []eos.PermissionLevel{
			{
				Actor:      eos.AN(payerAccount),
				Permission: eos.PN("active"),
			},
			{
				Actor:      eos.AN(fromAccount),
				Permission: eos.PN("active"),
			},
		},
		ActionData: eos.NewActionData(token.Transfer{
			From:     eos.AN(fromAccount),
			To:       eos.AN(toAccount),
			Quantity: asset,
			Memo:     memo,
		}),
	}

	// 2. Create transaction
	txOpts := &eos.TxOptions{}
	if err := txOpts.FillFromChain(ctx, api); err != nil {
		return nil, err
	}

	tx := eos.NewTransaction([]*eos.Action{action}, txOpts)

	// 3. User signature
	userKey, err := ecc.NewPrivateKey(userPrivateKeyWIF)
	if err != nil {
		return nil, err
	}
	api.SetSigner(eos.NewKeyBag())
	api.Signer.ImportPrivateKey(ctx, userPrivateKeyWIF)

	stx := eos.NewSignedTransaction(tx)

	signedTx, err := api.Signer.Sign(ctx, stx, txOpts.ChainID, userKey.PublicKey())
	if err != nil {
		return nil, err
	}


	return signedTx, nil
}

// Sign and broadcast transaction by payer account
func SignAndBroadcastByPayer(
	ctx context.Context,
	api *eos.API,
	tx *eos.SignedTransaction,
	payerPrivateKey string,
) (*eos.PushTransactionFullResp, error) {
	payerKey, err := ecc.NewPrivateKey(payerPrivateKey)
	if err != nil {
		return nil, err
	}

	api.SetSigner(eos.NewKeyBag())
	api.Signer.ImportPrivateKey(ctx, payerPrivateKey)

	txOpts := &eos.TxOptions{}
	if err := txOpts.FillFromChain(ctx, api); err != nil {
		return nil, err
	}

	signedTx, err := api.Signer.Sign(ctx, tx, []byte(txOpts.ChainID), payerKey.PublicKey())
	if err != nil {
		return nil, err
	}

	// Reverse signatures - payer's signature must be first
	if len(signedTx.Signatures) > 1 {
		signedTx.Signatures = append(signedTx.Signatures[1:], signedTx.Signatures[0])
	}

	packed, err := signedTx.Pack(eos.CompressionNone)
	if err != nil {
		return nil, err
	}

	resp, err := api.PushTransaction(ctx, packed)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
