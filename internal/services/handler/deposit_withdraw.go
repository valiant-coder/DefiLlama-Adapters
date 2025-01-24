package handler

import (
	"context"
	"encoding/json"
	"exapp-go/internal/db/db"
	"exapp-go/pkg/hyperion"
	"exapp-go/pkg/utils"
	"log"

	eosgo "github.com/eoscanada/eos-go"
	"github.com/shopspring/decimal"
)

/*
	enum AssetType {
	  NATIVE = 1,
	  BTC = 2,
	  EXSAT = 3,
	  EXSAT_EVM = 4,
	};
*/
func (s *Service) handleDeposit(action hyperion.Action) error {
	ctx := context.Background()
	var data struct {
		Account   string `json:"account"`
		Contract  string `json:"contract"`
		Quantity  string `json:"quantity"`
		AssetType uint8  `json:"asset_type"`
		Fee       string `json:"fee"`
	}
	if err := json.Unmarshal(action.Act.Data, &data); err != nil {
		log.Printf("Unmarshal deposit failed: %v", err)
		return nil
	}

	if data.AssetType == 3 {
		log.Printf("Asset type is exsat, skip")
		return nil
	}

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
	depositTime, err := utils.ParseTime(action.Timestamp)
	if err != nil {
		log.Printf("parse action time failed: %v", err)
		return nil
	}
	var chianName string
	if data.AssetType == 1 {
		chianName = "eos"
	} else {
		chianName = "exsat"
	}
	record := &db.DepositRecord{
		Symbol:         asset.Symbol.Symbol,
		UID:            uid,
		Amount:         decimal.New(int64(asset.Amount), -int32(asset.Symbol.Precision)),
		Fee:            decimal.New(int64(feeAsset.Amount), -int32(feeAsset.Symbol.Precision)),
		Status:         db.DepositStatusSuccess,
		Time:           depositTime,
		TxHash:         action.TrxID,
		SourceTxID:     action.TrxID,
		DepositAddress: "",
		ChainName:      chianName,
		BlockNumber:    uint64(action.BlockNum),
	}
	err = s.repo.UpsertDepositRecord(ctx, record)
	if err != nil {
		log.Printf("Create deposit record failed: %v-%v", data, err)
		return nil
	}

	return nil

}

func (s *Service) handleEOSSend(action hyperion.Action) error {
	ctx := context.Background()
	var data struct {
		From     string `json:"from"`
		To       string `json:"to"`
		Contract string `json:"contract"`
		Quantity string `json:"quantity"`
		Fee      string `json:"fee"`
	}
	if err := json.Unmarshal(action.Act.Data, &data); err != nil {
		log.Printf("Unmarshal withdraw failed: %v", err)
		return nil
	}

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

	uid, err := s.repo.GetUIDByEOSAccount(ctx, data.From)
	if err != nil {
		log.Printf("Get uid by eos account failed: %v-%v", data, err)
		return nil
	}

	withdrawTime, err := utils.ParseTime(action.Timestamp)
	if err != nil {
		log.Printf("parse action time failed: %v", err)
		return nil
	}

	record := &db.WithdrawRecord{
		UID:         uid,
		Symbol:      asset.Symbol.Symbol,
		ChainName:   "eos",
		Amount:      decimal.New(int64(asset.Amount), -int32(asset.Symbol.Precision)),
		Fee:         decimal.New(int64(feeAsset.Amount), -int32(feeAsset.Symbol.Precision)),
		BridgeFee:   decimal.NewFromInt(0),
		Status:      db.WithdrawStatusSuccess,
		TxHash:      action.TrxID,
		WithdrawAt:  withdrawTime,
		BlockNumber: action.BlockNum,
		Recipient:   data.To,
	}
	err = s.repo.CreateWithdrawRecord(ctx, record)
	if err != nil {
		log.Printf("Create withdraw record failed: %v-%v", data, err)
		return nil
	}
	return nil
}
