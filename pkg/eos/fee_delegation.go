package eos

import (
	"context"
	"encoding/hex"
	"log"

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

func decodeSignedTransaction(signedTx string) (*eos.SignedTransaction, error) {
	txBytes, err := hex.DecodeString(signedTx)
	if err != nil {
		return nil, err
	}
	tx := &eos.SignedTransaction{}
	err = eos.UnmarshalBinary(txBytes, tx)
	if err != nil {
		return nil, err
	}
	return tx, nil
}

func CheckIsLegitTransaction(signedTx string, payEosAccount, eosAccount string, permission string) bool {
	tx, err := decodeSignedTransaction(signedTx)
	if err != nil {
		log.Printf("decode signed transaction failed: %v", err)
		return false
	}

	if len(tx.Actions) == 0 {
		log.Printf("transaction actions is empty")
		return false
	}

	
	// first action must be pay account noop
	if tx.Actions[0].Account != eos.AN(payEosAccount) || tx.Actions[0].Name != eos.ActN("noop") {
		log.Printf("transaction first action is not payer noop")
		return false
	}

	for _, action := range tx.Actions[1:] {
		for _, authorization := range action.Authorization {
			if authorization.Actor != eos.AN(eosAccount) || authorization.Permission != eos.PN(permission) {
				log.Printf("transaction authorization is not legit, actor: %s, permission: %s", authorization.Actor, authorization.Permission)
				return false
			}
		}
	}

	return true
}

// Sign and broadcast transaction by payer account
func SignAndBroadcastByPayer(
	ctx context.Context,
	api *eos.API,
	singleSignedTx string,
	payerPrivateKey string,
) (*eos.PushTransactionFullResp, error) {

	tx, err := decodeSignedTransaction(singleSignedTx)
	if err != nil {
		return nil, err
	}

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

	fullSignedTx, err := api.Signer.Sign(ctx, tx, []byte(txOpts.ChainID), payerKey.PublicKey())
	if err != nil {
		log.Printf("sign transaction failed: %v", err)
		return nil, err
	}

	// Reverse signatures - payer's signature must be first
	if len(fullSignedTx.Signatures) > 1 {
		fullSignedTx.Signatures = append(fullSignedTx.Signatures[1:], fullSignedTx.Signatures[0])
	}

	fullSignedTxBytes, err := eos.MarshalBinary(fullSignedTx)
	if err != nil {
		log.Fatalf("Failed to marshal transaction: %v", err)
	}

	log.Printf("fullSignedTx: %v", hex.EncodeToString(fullSignedTxBytes))

	packed, err := fullSignedTx.Pack(eos.CompressionNone)
	if err != nil {
		log.Printf("pack transaction failed: %v", err)
		return nil, err
	}

	resp, err := api.PushTransaction(ctx, packed)
	if err != nil {
		log.Printf("push transaction failed: %v", err)
		return nil, err
	}
	return resp, nil
}
