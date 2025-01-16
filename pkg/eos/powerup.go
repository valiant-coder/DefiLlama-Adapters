package eos

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"

	"github.com/eoscanada/eos-go"
)

func PowerUp(
	ctx context.Context,
	endpoint,
	accountName,
	privKey string,
	cpuEOS uint64,
	netEOS uint64,
) error {
	// Get weight values
	netWeight, cpuWeight, err := GetPowerUpStatusWeight(ctx, endpoint)
	if err != nil {
		return fmt.Errorf("get powerup status weight error: %w", err)
	}

	// Calculate netFrac and cpuFrac
	netFrac, err := calculateFrac(netEOS, netWeight)
	if err != nil {
		return fmt.Errorf("calculate net frac error: %w", err)
	}

	cpuFrac, err := calculateFrac(cpuEOS, cpuWeight)
	if err != nil {
		return fmt.Errorf("calculate cpu frac error: %w", err)
	}

	api := eos.New(endpoint)
	keyBag := &eos.KeyBag{}
	err = keyBag.ImportPrivateKey(ctx, privKey)
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
			MaxPayment: eos.NewEOSAsset(int64(100000)),
		}),
	}

	_, err = api.SignPushActions(ctx, action)
	if err != nil {
		return fmt.Errorf("send powerup transaction error: %w", err)
	}

	return nil
}

// calculateFrac calculates the frac value
func calculateFrac(payment uint64, weight string) (uint64, error) {
	// Remove quotes from weight string
	weight = strings.Trim(weight, "\"")

	// Create big.Int objects
	weightBig := new(big.Int)
	paymentBig := new(big.Int).SetUint64(payment)
	pow15 := new(big.Int).Exp(big.NewInt(10), big.NewInt(15), nil)

	_, ok := weightBig.SetString(weight, 10)
	if !ok {
		return 0, fmt.Errorf("invalid weight value: %s", weight)
	}

	// Calculate payment * 10^15
	numerator := new(big.Int).Mul(paymentBig, pow15)

	// Calculate (payment * 10^15) / weight
	frac := new(big.Int).Div(numerator, weightBig)

	// Check if result exceeds uint64 range
	if !frac.IsUint64() {
		return 0, fmt.Errorf("calculated frac exceeds uint64 range")
	}

	return frac.Uint64(), nil
}

type PowerUpArgs struct {
	Payer      eos.AccountName `eos:"payer"`
	Receiver   eos.AccountName `eos:"receiver"`
	Days       uint32          `eos:"days"`
	NetFrac    uint64          `eos:"net_frac"`
	CPUFrac    uint64          `eos:"cpu_frac"`
	MaxPayment eos.Asset       `eos:"max_payment"`
}

type PowerUpStatus struct {
	Version     uint64        `json:"version"`
	Net         PowerUpWeight `json:"net"`
	CPU         PowerUpWeight `json:"cpu"`
	PowerUpDays uint64        `json:"powerup_days"`
}

type PowerUpWeight struct {
	Weight string `json:"weight"`
}

func GetPowerUpStatusWeight(ctx context.Context, endpoint string) (string, string, error) {
	api := eos.New(endpoint)
	resp, err := api.GetTableRows(ctx, eos.GetTableRowsRequest{
		Code:       "eosio",
		Scope:      "0",
		Table:      "powup.state",
		JSON:       true,
		LowerBound: "0",
		UpperBound: "-1",
		Limit:      1000,
	})
	if err != nil {
		return "", "", err
	}
	var rows []PowerUpStatus
	err = json.Unmarshal(resp.Rows, &rows)
	if err != nil {
		return "", "", err
	}

	if len(rows) == 0 {
		return "", "", fmt.Errorf("no powerup status found")
	}

	return rows[0].Net.Weight, rows[0].CPU.Weight, nil
}
