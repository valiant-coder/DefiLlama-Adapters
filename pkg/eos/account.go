package eos

import (
	"context"
	"fmt"

	"github.com/eoscanada/eos-go"
	"github.com/eoscanada/eos-go/ecc"
	"github.com/eoscanada/eos-go/system"
)

func NewAccountManager(config *Config) (*AccountManager, error) {
	api := eos.New(config.URL)
	api.SetSigner(eos.NewKeyBag())
	api.Signer.ImportPrivateKey(context.Background(), config.CreatorPrivKey)

	return &AccountManager{
		eosClient: api,
		config:    config,
	}, nil
}


// GenerateAccount creates an EOS account with 2/2 multisig permissions
func (am *AccountManager) GenerateAccount(ctx context.Context, accountName string) (*AccountKeys, error) {
	// Generate keypairs for social and hash shards
	socialPrivKey, err := ecc.NewRandomPrivateKey()
	if err != nil {
		return nil, fmt.Errorf("generate social key error: %w", err)
	}
	socialPubKey := socialPrivKey.PublicKey().String()

	// Generate hash shard keypair
	hashPrivKey, err := ecc.NewRandomPrivateKey()
	if err != nil {
		return nil, fmt.Errorf("generate hash key error: %w", err)
	}
	hashPubKey := hashPrivKey.PublicKey().String()

	// Create multisig authority structure
	authority := create2MultiSigAuthority(socialPubKey, hashPubKey)

	// Create actions for account creation, RAM purchase, and resource delegation
	createAccountAction := am.createAccountAction(accountName, authority)
	buyRamAction := am.createBuyRamAction(accountName)
	delegateBwAction := am.createDelegateBwAction(accountName)

	actions := []*eos.Action{
		createAccountAction,
		buyRamAction,
		delegateBwAction,
	}

	_, err = am.eosClient.SignPushActions(ctx, actions...)
	if err != nil {
		return nil, fmt.Errorf("sign and push actions error: %w", err)
	}

	keys := &AccountKeys{
		SocialPrivKey: socialPrivKey.String(),
		HashPrivKey:   hashPrivKey.String(),
	}

	return keys, nil
}

// UpdateMultiSigPermission updates account permissions to 2/3 multisig
func (am *AccountManager) UpdateMultiSigPermission(ctx context.Context, accountName string, devicePubKey, recoveryPubKey string, keys *AccountKeys) error {
	socialPrivKey, err := ecc.NewPrivateKey(keys.SocialPrivKey)
	if err != nil {
		return fmt.Errorf("parse social private key error: %w", err)
	}
	socialPubKey := socialPrivKey.PublicKey().String()

	authority := create3MultiSigAuthority(socialPubKey, devicePubKey, recoveryPubKey)

	updateOwnerAuthAction := system.NewUpdateAuth(eos.AN(accountName), eos.PN("owner"), eos.PN(""), authority, eos.PN("owner"))
	updateActiveAuthAction := system.NewUpdateAuth(eos.AN(accountName), eos.PN("active"), eos.PN("owner"), authority, eos.PN("owner"))
	am.eosClient.Signer.ImportPrivateKey(ctx, keys.SocialPrivKey)
	am.eosClient.Signer.ImportPrivateKey(ctx, keys.HashPrivKey)

	_, err = am.eosClient.SignPushActions(ctx, updateOwnerAuthAction, updateActiveAuthAction)
	if err != nil {
		return fmt.Errorf("sign and push update auth actions error: %w", err)
	}
	return nil
}

// CreateDeviceRecoveryTx prepares unsigned tx for updating account permissions
// Requires both recovery and social key signatures
func (am *AccountManager) CreateDeviceRecoveryTx(
	ctx context.Context,
	accountName string,
	newDevicePubKey string,
	socialPubKey string,
	recoveryPubKey string,
) (*eos.SignedTransaction, error) {
	authority := create3MultiSigAuthority(socialPubKey, newDevicePubKey, recoveryPubKey)

	updateOwnerAuthAction := system.NewUpdateAuth(
		eos.AN(accountName),
		eos.PN("owner"),
		eos.PN(""),
		authority,
		eos.PN("owner"),
	)
	updateActiveAuthAction := system.NewUpdateAuth(
		eos.AN(accountName),
		eos.PN("active"),
		eos.PN("owner"),
		authority,
		eos.PN("owner"),
	)

	tx := eos.NewSignedTransaction(eos.NewTransaction([]*eos.Action{
		updateOwnerAuthAction,
		updateActiveAuthAction,
	}, nil))

	return tx, nil
}

// SignAndPushRecoveryTx validates recovery signature, signs with social key and broadcasts
func (am *AccountManager) SignAndPushRecoveryTx(
	ctx context.Context,
	signedTx *eos.SignedTransaction,
	accountName string,
	newDevicePubKey string,
	socialPrivKey string,
	recoveryPubKey string,
) error {
	if len(signedTx.Actions) != 2 {
		return fmt.Errorf("invalid transaction: expected 2 actions, got %d", len(signedTx.Actions))
	}

	recoveryKey, err := ecc.NewPublicKey(recoveryPubKey)
	if err != nil {
		return fmt.Errorf("parse recovery public key error: %w", err)
	}

	hasRecoverySignature := false
	pubKeys, err := signedTx.SignedByKeys([]byte(am.config.ChainID))
	if err != nil {
		return fmt.Errorf("get signed by keys error: %w", err)
	}
	for _, pubKey := range pubKeys {
		if pubKey.String() == recoveryKey.String() {
			hasRecoverySignature = true
			break
		}
	}

	if !hasRecoverySignature {
		return fmt.Errorf("transaction missing recovery key signature")
	}

	for _, action := range signedTx.Actions {
		if action.Account != eos.AN("eosio") || action.Name != eos.ActN("updateauth") {
			return fmt.Errorf("invalid action: expected eosio:updateauth")
		}
		
		auth, ok := action.ActionData.Data.(system.UpdateAuth)
		if !ok {
			return fmt.Errorf("invalid action data type")
		}

		if auth.Account != eos.AN(accountName) {
			return fmt.Errorf("invalid account name in permission update")
		}

		authority := auth.Auth
	
		if authority.Threshold != 2 || len(authority.Keys) != 3 {
			return fmt.Errorf("invalid multisig configuration")
		}

		hasNewDeviceKey := false
		for _, key := range authority.Keys {
			if key.PublicKey.String() == newDevicePubKey {
				hasNewDeviceKey = true
				break
			}
		}
		if !hasNewDeviceKey {
			return fmt.Errorf("new device public key not found in authority")
		}
	}

	am.eosClient.Signer.ImportPrivateKey(ctx, socialPrivKey)

	sp, err := ecc.NewPrivateKey(socialPrivKey)
	if err != nil {
		return fmt.Errorf("parse social private key error: %w", err)
	}

	signedTx, err = am.eosClient.Signer.Sign(ctx, signedTx, []byte(am.config.ChainID),sp.PublicKey())
	if err != nil {
		return fmt.Errorf("sign transaction with social key error: %w", err)
	}


	packed, err := signedTx.Pack(eos.CompressionNone)
	if err != nil {
		return fmt.Errorf("pack transaction error: %w", err)
	}

	_, err = am.eosClient.PushTransaction(ctx, packed)
	if err != nil {
		return fmt.Errorf("push signed recovery transaction error: %w", err)
	}

	return nil
}

// create2MultiSigAuthority builds 2/2 multisig authority
func create2MultiSigAuthority(socialPubKey, hashPubKey string) eos.Authority {
	return eos.Authority{
		Threshold: 2,
		Keys: []eos.KeyWeight{
			{
				PublicKey: toPubKey(socialPubKey),
				Weight:    1,
			},
			{
				PublicKey: toPubKey(hashPubKey),
				Weight:    1,
			},
		},
		Accounts: []eos.PermissionLevelWeight{},
		Waits:    []eos.WaitWeight{},
	}
}

// create3MultiSigAuthority builds 2/3 multisig authority
func create3MultiSigAuthority(socialPubKey, devicePubKey, recoveryPubKey string) eos.Authority {
	return eos.Authority{
		Threshold: 2,
		Keys: []eos.KeyWeight{
			{
				PublicKey: toPubKey(socialPubKey),
				Weight:    1,
			},
			{
				PublicKey: toPubKey(devicePubKey),
				Weight:    1,
			},
			{
				PublicKey: toPubKey(recoveryPubKey),
				Weight:    1,
			},
		},
	}
}

// createAccountAction builds new account creation action
func (am *AccountManager) createAccountAction(accountName string, authority eos.Authority) *eos.Action {
	return &eos.Action{
		Account: eos.AN("eosio"),
		Name:    eos.ActN("newaccount"),
		Authorization: []eos.PermissionLevel{
			{Actor: eos.AN(am.config.CreatorAccount), Permission: eos.PN("active")},
		},
		ActionData: eos.NewActionData(system.NewAccount{
			Creator: eos.AN(am.config.CreatorAccount),
			Name:    eos.AN(accountName),
			Owner:   authority,
			Active:  authority,
		}),
	}
}

// createBuyRamAction builds RAM purchase action
func (am *AccountManager) createBuyRamAction(accountName string) *eos.Action {
	return &eos.Action{
		Account: eos.AN("eosio"),
		Name:    eos.ActN("buyrambytes"),
		Authorization: []eos.PermissionLevel{
			{Actor: eos.AN(am.config.CreatorAccount), Permission: eos.PN("active")},
		},
		ActionData: eos.NewActionData(system.BuyRAMBytes{
			Payer:    eos.AN(am.config.CreatorAccount),
			Receiver: eos.AN(accountName),
			Bytes:    8192, // Adjust RAM size as needed
		}),
	}
}

// createDelegateBwAction builds CPU/NET bandwidth delegation action
func (am *AccountManager) createDelegateBwAction(accountName string) *eos.Action {
	return &eos.Action{
		Account: eos.AN("eosio"),
		Name:    eos.ActN("delegatebw"),
		Authorization: []eos.PermissionLevel{
			{Actor: eos.AN(am.config.CreatorAccount), Permission: eos.PN("active")},
		},
		ActionData: eos.NewActionData(system.DelegateBW{
			From:     eos.AN(am.config.CreatorAccount),
			Receiver: eos.AN(accountName),
			StakeNet: eos.NewEOSAsset(1000), // 1.0 EOS for NET
			StakeCPU: eos.NewEOSAsset(1000), // 1.0 EOS for CPU
			Transfer: false,
		}),
	}
}

// toPubKey converts string to PublicKey, panics on error
func toPubKey(pubKeyStr string) ecc.PublicKey {
	pubKey, err := ecc.NewPublicKey(pubKeyStr)
	if err != nil {
		panic(err)
	}
	return pubKey
}
