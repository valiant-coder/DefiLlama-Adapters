package handler

import (
	"context"
	"encoding/json"
	"errors"
	"exapp-go/internal/db/db"
	"exapp-go/pkg/hyperion"
	"log"
	"time"

	eosgo "github.com/eoscanada/eos-go"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)





func (s *Service) handleBridgeDeposit(action hyperion.Action) error {
	ctx := context.Background()
	var data struct {
		PermissionID     string `json:"permission_id"`
		GlobalStatus     uint8  `json:"global_status"`
		Applicant        string `json:"applicant"`
		ChainName        string `json:"chain_name"`
		SourceContract   string `json:"source_contract"`
		SourceSymbol     string `json:"source_symbol"`
		DestContract     string `json:"dest_contract"`
		DestSymbol       string `json:"dest_symbol"`
		SenderAddress    string `json:"sender_address"`
		DepositAddress   string `json:"deposit_address"`
		RecipientAddress string `json:"recipient_address"`
		BlockHeight      string `json:"block_height"`
		TxID             string `json:"tx_id"`
		DepositAmount    string `json:"deposit_amount"`
		DepositFee       string `json:"deposit_fee"`
		TransferAmount   string `json:"transfer_amount"`
		TxTimestamp      string `json:"tx_timestamp"`
		CreateTimestamp  string `json:"create_timestamp"`
	}
	if err := json.Unmarshal(action.Act.Data, &data); err != nil {
		log.Printf("Unmarshal deposit failed: %v", err)
		return nil
	}

	if data.Applicant != s.exappCfg.AssetContract {
		log.Printf("Applicant is not %s, skip", s.exappCfg.AssetContract)
		return nil
	}

	depositAddress, err := s.repo.GetUserDepositAddressByAddress(ctx, data.DepositAddress)
	if err != nil {
		log.Printf("not found deposit address: %v-%v", data.DepositAddress, err)
		return nil
	}

	token, err := s.repo.GetToken(ctx, data.DestSymbol, data.ChainName)
	if err != nil {
		log.Printf("not found token: %v-%v", data.DestSymbol, err)
		return nil
	}

	depositAmount := decimal.RequireFromString(data.DepositAmount).Shift(-int32(token.Decimals))
	depositFee := decimal.RequireFromString(data.DepositFee).Shift(-int32(token.Decimals))

	record, err := s.repo.GetDepositRecordBySourceTxID(ctx, data.TxID)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("Get deposit record by source tx id failed: %v-%v", data, err)
			return nil
		} else {
			record = &db.DepositRecord{
				Symbol:         data.DestSymbol,
				UID:            depositAddress.UID,
				Amount:         depositAmount,
				Fee:            depositFee,
				Status:         db.DepositStatus(data.GlobalStatus),
				Time:           time.Now(),
				TxHash:         action.TrxID,
				SourceTxID:     data.TxID,
				DepositAddress: data.DepositAddress,
				ChainName:      data.ChainName,
			}
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
