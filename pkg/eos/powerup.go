package eos

import (
	"context"
	"fmt"

	"github.com/eoscanada/eos-go"
)

func PowerUp(
	ctx context.Context,
	endpoint,
	accountName,
	privKey string,
	netFrac,
	cpuFrac,
	maxPayment uint64,
) error {
	api := eos.New(endpoint)
	keyBag := &eos.KeyBag{}
	err := keyBag.ImportPrivateKey(ctx, privKey)
	if err != nil {
		return fmt.Errorf("import private key error: %w", err)
	}
	api.SetSigner(keyBag)

	action := &eos.Action{
		Account: eos.AN("eosio"),
		Name:    eos.ActN("powerup"),
		Authorization: []eos.PermissionLevel{
			{Actor: eos.AN(accountName), Permission: eos.PN("active")},
		},
		ActionData: eos.NewActionData(PowerUpArgs{
			Payer:      eos.AN(accountName),
			Receiver:   eos.AN(accountName),
			Days:       1,
			NetFrac:    netFrac,
			CPUFrac:    cpuFrac,
			MaxPayment: eos.NewEOSAsset(int64(maxPayment)),
		}),
	}

	_, err = api.SignPushActions(ctx, action)
	if err != nil {
		return fmt.Errorf("send powerup transaction error: %w", err)
	}

	return nil
}

type PowerUpArgs struct {
	Payer      eos.AccountName `eos:"payer"`
	Receiver   eos.AccountName `eos:"receiver"`
	Days       uint32          `eos:"days"`
	NetFrac    uint64          `eos:"net_frac"`
	CPUFrac    uint64          `eos:"cpu_frac"`
	MaxPayment eos.Asset       `eos:"max_payment"`
}
