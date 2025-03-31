package entity

import "exapp-go/internal/db/db"

type ReqDeposit struct {
	Symbol  string `json:"symbol" binding:"required"`
	ChainID uint8  `json:"chain_id" binding:"required"`
}

type RespDeposit struct {
	Address string `json:"address"`
}

type RespDepositRecord struct {
	ID             uint64 `json:"id"`
	Symbol         string `json:"symbol"`
	Amount         string `json:"amount"`
	ChainName      string `json:"chain_name"`
	SourceTxID     string `json:"source_tx_id"`
	DepositAddress string `json:"deposit_address"`
	// 0 pending 1success 2 fail
	Status    uint8 `json:"status"`
	DepositAt Time  `json:"deposit_at"`
}

func FormatDepositRecord(record *db.DepositRecord) RespDepositRecord {
	return RespDepositRecord{
		ID:             uint64(record.ID),
		Symbol:         record.Symbol,
		Amount:         record.Amount.String(),
		ChainName:      record.ChainName,
		SourceTxID:     record.SourceTxID,
		DepositAddress: record.DepositAddress,
		Status:         uint8(record.Status),
		DepositAt:      Time(record.Time),
	}
}

type RespWithdrawRecord struct {
	ID        uint64 `json:"id"`
	Symbol    string `json:"symbol"`
	Amount    string `json:"amount"`
	ChainName string `json:"chain_name"`
	Fee       string `json:"fee"`
	// 0 pending 1success 2 fail
	Status      uint8  `json:"status"`
	SendTxID    string `json:"send_tx_id"`
	WithdrawAt  Time   `json:"withdraw_at"`
	CompletedAt Time   `json:"completed_at"`
	Recipient   string `json:"recipient"`
}

func FormatWithdrawRecord(record *db.WithdrawRecord) RespWithdrawRecord {
	return RespWithdrawRecord{
		ID:          uint64(record.ID),
		Symbol:      record.Symbol,
		Amount:      record.Amount.String(),
		ChainName:   record.ChainName,
		Fee:         record.Fee.Add(record.BridgeFee).String(),
		Status:      uint8(record.Status),
		SendTxID:    record.SendTxID,
		WithdrawAt:  Time(record.WithdrawAt),
		CompletedAt: Time(record.CompletedAt),
		Recipient:   record.Recipient,
	}
}
