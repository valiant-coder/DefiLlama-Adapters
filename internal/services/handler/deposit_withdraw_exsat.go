package handler

import (
	"context"
	"encoding/json"
	"errors"
	"exapp-go/internal/db/db"
	"exapp-go/pkg/hyperion"
	"exapp-go/pkg/utils"
	"log"
	"strings"
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

	if data.Applicant != s.oneDexCfg.PortalContract {
		log.Printf("Applicant is not %s, skip", s.oneDexCfg.PortalContract)
		return nil
	}

	depositAddress, err := s.repo.GetUserDepositAddressByAddress(ctx, data.DepositAddress)
	if err != nil {
		log.Printf("not found deposit address: %v-%v", data.DepositAddress, err)
		return nil
	}

	symbol := strings.TrimSuffix(data.DestSymbol, ".t")

	token, err := s.repo.GetToken(ctx, symbol)
	if err != nil {
		log.Printf("not found token: %v-%v", symbol, err)
		return nil
	}

	depositAmount := decimal.RequireFromString(data.DepositAmount).Shift(-int32(token.ExsatDecimals))
	depositFee := decimal.RequireFromString(data.DepositFee).Shift(-int32(token.ExsatDecimals))

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
		} else {
			record = &db.DepositRecord{
				Symbol:         symbol,
				UID:            depositAddress.UID,
				Amount:         depositAmount,
				Fee:            depositFee,
				Status:         db.DepositStatus(data.GlobalStatus),
				Time:           depositTime,
				TxHash:         action.TrxID,
				SourceTxID:     data.TxID,
				DepositAddress: data.DepositAddress,
				ChainName:      data.ChainName,
				BlockNumber:    uint64(action.BlockNum),
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
	withdrawAt, err := utils.ParseTime(action.Timestamp)
	if err != nil {
		log.Printf("Parse withdraw timestamp failed: %v-%v", data, err)
		return nil
	}

	token, err := s.repo.GetToken(ctx, asset.Symbol.Symbol)
	if err != nil {
		log.Printf("Get token failed: %v-%v", data, err)
		return nil
	}
	var targetChain db.ChainInfo
	for _, chain := range token.Chains {
		if chain.ChainName == data.ChainName {
			targetChain = chain
			break
		}
	}

	var withdrawStatus db.WithdrawStatus
	var sendTxID string
	var completedAt time.Time
	if targetChain.ChainName == "exsat" {
		withdrawStatus = db.WithdrawStatusSuccess
		sendTxID, err = s.hyperionCli.GetEvmTxIDByEosTxID(action.TrxID)
		if err != nil {
			log.Printf("Get evm tx id by eos tx id failed: %v-%v", action.TrxID, err)
			sendTxID = action.TrxID
		}
		completedAt = withdrawAt
	} else if targetChain.ChainName == "eos" {
		withdrawStatus = db.WithdrawStatusSuccess
		sendTxID = action.TrxID
		completedAt = withdrawAt
	} else {
		withdrawStatus = db.WithdrawStatusPending
	}

	err = s.repo.CreateWithdrawRecord(ctx, &db.WithdrawRecord{
		UID:         uid,
		Symbol:      asset.Symbol.Symbol,
		ChainName:   data.ChainName,
		Amount:      decimal.New(int64(asset.Amount), -int32(asset.Symbol.Precision)),
		Fee:         decimal.New(int64(feeAsset.Amount), -int32(feeAsset.Symbol.Precision)),
		BridgeFee:   targetChain.ExsatWithdrawFee,
		Status:      withdrawStatus,
		SendTxID:    sendTxID,
		TxHash:      action.TrxID,
		WithdrawAt:  withdrawAt,
		CompletedAt: completedAt,
		BlockNumber: action.BlockNum,
		Recipient:   data.Recipient,
	})
	if err != nil {
		log.Printf("Create withdraw record failed: %v-%v", data, err)
		return nil
	}

	go s.updateUserTokenBalance(data.Account, "active")

	return nil
}

func (s *Service) updateWithdraw(action hyperion.Action) error {
	var data struct {
		GlobalStatus uint8 `json:"global_status"`
		// target send tx id
		TxID           string `json:"tx_id"`
		WithdrawAmount string `json:"withdraw_amount"`
		WithdrawFee    string `json:"withdraw_fee"`
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

	token, err := s.repo.GetToken(ctx, record.Symbol)
	if err != nil {
		log.Printf("Get token failed: %v-%v", data, err)
		return nil
	}
	var targetChain db.ChainInfo
	for _, chain := range token.Chains {
		if chain.ChainName == record.ChainName {
			targetChain = chain
			break
		}
	}
	bridgeFee := decimal.RequireFromString(data.WithdrawFee).Shift(-int32(targetChain.ExsatTokenDecimals))

	record.Status = db.WithdrawStatus(data.GlobalStatus)
	record.CompletedAt = completedAt
	record.BridgeFee = bridgeFee
	record.SendTxID = data.TxID
	err = s.repo.UpdateWithdrawRecord(ctx, record)
	if err != nil {
		log.Printf("Update withdraw record failed: %v-%v", data, err)
		return nil
	}

	return nil

}
