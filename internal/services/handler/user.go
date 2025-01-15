package handler

import (
	"context"
	"encoding/json"
	"exapp-go/pkg/hyperion"
	"log"

	"exapp-go/internal/db/db"

	eosgo "github.com/eoscanada/eos-go"
	"github.com/shopspring/decimal"
)

func (s *Service) handleNewAccount(action hyperion.Action) error {
	var data struct {
		ID      string `json:"id"`
		Account string `json:"account"`
		Pubkey  string `json:"pubkey"`
	}
	if err := json.Unmarshal(action.Act.Data, &data); err != nil {
		log.Printf("Unmarshal new account failed: %v", err)
		return nil
	}

	ctx := context.Background()
	credential, err := s.repo.GetUserCredentialByPubkey(ctx, data.Pubkey)
	if err != nil {
		log.Printf("Get user credential by pubkey failed: %v", err)
		return nil
	}

	credential.EOSAccount = data.Account
	credential.EOSPermissions = "active,owner"
	err = s.repo.UpdateUserCredential(ctx, credential)
	if err != nil {
		log.Printf("Update user credential failed: %v-%v", data, err)
		return nil
	}

	return nil
}

func (s *Service) handleDeposit(action hyperion.Action) error {
	var data struct {
		Account  string `json:"account"`
		Contract string `json:"contract"`
		Quantity string `json:"quantity"`
	}
	if err := json.Unmarshal(action.Act.Data, &data); err != nil {
		log.Printf("Unmarshal deposit failed: %v", err)
		return nil
	}

	ctx := context.Background()
	uid, err := s.repo.GetUIDByEOSAccount(ctx, data.Account)
	if err != nil {
		log.Printf("Get uid by eos account failed: %v-%v", data, err)
		return nil
	}

	asset, err := eosgo.NewAssetFromString(data.Quantity)
	if err != nil {
		log.Printf("New asset from string failed: %v-%v", data, err)
		return nil
	}

	err = s.repo.CreateDepositRecord(ctx, &db.DepositRecord{
		Symbol: asset.Symbol.Symbol,
		UID:    uid,
		Amount: decimal.New(int64(asset.Amount), -int32(asset.Symbol.Precision)),
		Status: db.DepositStatusSuccess,
		TxHash: action.TrxID,
	})
	if err != nil {
		log.Printf("Create deposit record failed: %v-%v", data, err)
		return nil
	}

	go s.updateUserTokenBalance(data.Account)

	return nil
}

func (s *Service) updateUserTokenBalance(account string) error {
	go s.publisher.PublishBalanceUpdate(account)

	return nil
}
