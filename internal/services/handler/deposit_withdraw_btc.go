package handler

import (
	"context"
	"encoding/json"
	"errors"
	"exapp-go/internal/db/db"
	"exapp-go/pkg/hyperion"
	"exapp-go/pkg/utils"
	"log"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

func (s *Service) handleBTCDeposit(action hyperion.Action) error {
	ctx := context.Background()
	var data struct {
		Amount       string `json:"amount"`
		Fee          string `json:"fee"`
		TxID         string `json:"tx_id"`
		GlobalStatus int    `json:"global_status"`
		BTCAddress   string `json:"btc_address"`
	}
	if err := json.Unmarshal(action.Act.Data, &data); err != nil {
		log.Printf("Unmarshal withdraw failed: %v", err)
		return nil
	}

	depositAddress, err := s.repo.GetUserDepositAddressByAddress(ctx, data.BTCAddress)
	if err != nil {
		log.Printf("not found deposit address: %v-%v", data.BTCAddress, err)
		return nil
	}

	depositAmount := decimal.RequireFromString(data.Amount).Shift(-8)
	depositFee := decimal.RequireFromString(data.Fee).Shift(-8)

	depositTime, err := utils.ParseTime(action.Timestamp)
	if err != nil {
		log.Printf("parse action time failed: %v", err)
		return nil
	}

	record, err := s.repo.GetDepositRecordBySourceTxID(ctx, data.TxID)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("Get deposit record by source tx id failed: %v-%v", data, err)
			return nil
		}
		record = &db.DepositRecord{
			Symbol:         "BTC",
			UID:            depositAddress.UID,
			Amount:         depositAmount,
			Fee:            depositFee,
			Status:         db.DepositStatus(data.GlobalStatus),
			Time:           depositTime,
			TxHash:         action.TrxID,
			SourceTxID:     data.TxID,
			DepositAddress: data.BTCAddress,
			ChainName:      s.exsatCfg.BTCChainName,
			BlockNumber:    uint64(action.BlockNum),
		}
	} else {
		record.Status = db.DepositStatus(data.GlobalStatus)
		record.TxHash = action.TrxID
	}

	err = s.repo.UpsertDepositRecord(ctx, record)
	if err != nil {
		log.Printf("Create deposit record failed: %v-%v", data, err)
		return nil
	}

	return nil

}



func (s *Service) updateBTCWithdraw(action hyperion.Action) error {
	var data struct {
		GlobalStatus uint8 `json:"global_status"`
		// target send tx id
		TxID           string `json:"tx_id"`
		Amount         string `json:"amount"`
		Fee            string `json:"fee"`
		TransactionID  string `json:"transaction_id"`
	}
	if err := json.Unmarshal(action.Act.Data, &data); err != nil {
		log.Printf("Unmarshal withdraw failed: %v", err)
		return nil
	}
	ctx := context.Background()

	record, err := s.repo.GetWithdrawRecordByTxHash(ctx, data.TransactionID)
	if err != nil {
		log.Printf("Get withdraw record by tx hash failed: %v-%v", data, err)
		return nil
	}

	completedAt, err := utils.ParseTime(action.Timestamp)
	if err != nil {
		log.Printf("Parse withdraw timestamp failed: %v-%v", data, err)
		return nil
	}


	bridgeFee := decimal.RequireFromString(data.Fee).Shift(-int32(8))

	record.Status = db.WithdrawStatus(data.GlobalStatus)
	record.CompletedAt = completedAt
	record.BridgeFee = bridgeFee
	record.SendTxID = data.TxID
	err = s.repo.UpdateWithdrawRecord(ctx, record)
	if err != nil {
		log.Printf("Update btc withdraw record failed: %v-%v", data, err)
		return nil
	}

	return nil

}
