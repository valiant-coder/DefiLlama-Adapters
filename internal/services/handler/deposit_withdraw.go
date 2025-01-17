package handler

import (
	"context"
	"encoding/json"
	"exapp-go/internal/db/db"
	"exapp-go/pkg/hyperion"
	"exapp-go/pkg/utils"
	"log"
	"time"

	eosgo "github.com/eoscanada/eos-go"
	"github.com/shopspring/decimal"
)

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

	depositTime, err := utils.ParseTime(action.Timestamp)
	if err != nil {
		log.Printf("parse action time failed: %v", err)
		return nil
	}
	err = s.repo.CreateDepositRecord(ctx, &db.DepositRecord{
		Symbol: asset.Symbol.Symbol,
		UID:    uid,
		Amount: decimal.New(int64(asset.Amount), -int32(asset.Symbol.Precision)),
		Status: db.DepositStatusSuccess,
		TxHash: action.TrxID,
		Time:   depositTime,
	})
	if err != nil {
		log.Printf("Create deposit record failed: %v-%v", data, err)
		return nil
	}

	go s.updateUserTokenBalance(data.Account)

	return nil
}

func (s *Service) handleWithdraw(action hyperion.Action) error {
	var data struct {
		Account      string `json:"account"`
		ChainName    string `json:"chain_name"`
		ChainType    string `json:"chain_type"`
		Contract     string `json:"contract"`
		Quantity     string `json:"quantity"`
		Fee          string `json:"fee"`
		Recipient    string `json:"recipient"`
		TokenAddress string `json:"token_address"`
	}
	if err := json.Unmarshal(action.Act.Data, &data); err != nil {
		log.Printf("Unmarshal withdraw failed: %v", err)
		return nil
	}

	ctx := context.Background()

	asset, err := eosgo.NewAssetFromString(data.Quantity)
	if err != nil {
		log.Printf("New asset from string failed: %v-%v", data, err)
		return nil
	}
	feeAsset, err := eosgo.NewAssetFromString(data.Fee)
	if err != nil {
		log.Printf("New asset from string failed: %v-%v", data, err)
		return nil
	}

	uid, err := s.repo.GetUIDByEOSAccount(ctx, data.Account)
	if err != nil {
		log.Printf("Get uid by eos account failed: %v-%v", data, err)
		return nil
	}

	err = s.repo.CreateWithdrawRecord(ctx, &db.UserWithdrawRecord{
		UID:       uid,
		Symbol:    asset.Symbol.Symbol,
		ChainName: data.ChainName,
		Amount:    decimal.New(int64(asset.Amount), -int32(asset.Symbol.Precision)),
		Fee:       decimal.New(int64(feeAsset.Amount), -int32(feeAsset.Symbol.Precision)),
		Status:    db.WithdrawStatusPending,
		TxHash:    action.TrxID,
		Time:      time.Now(),
	})
	if err != nil {
		log.Printf("Create withdraw record failed: %v-%v", data, err)
		return nil
	}

	go s.updateUserTokenBalance(data.Account)

	return nil
}



func (s *Service) updateWithdraw(action hyperion.Action) error {

	var data struct {
		
	}
	if err := json.Unmarshal(action.Act.Data, &data); err != nil {
		log.Printf("Unmarshal withdraw failed: %v", err)
		return nil
	}

	return nil


}
