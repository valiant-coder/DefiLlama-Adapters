package db

import (
	"context"
	"exapp-go/pkg/queryparams"
	"fmt"
	"strings"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func init() {
	addMigrateFunc(func(repo *Repo) error {
		return repo.AutoMigrate(
			&DepositRecord{},
			&UserDepositAddress{},
			&WithdrawRecord{},
		)
	})
}

type DepositStatus uint8

const (
	DepositStatusPending DepositStatus = iota
	DepositStatusSuccess
	DepositStatusFailed
)

type DepositRecord struct {
	gorm.Model
	Symbol         string          `gorm:"column:symbol;type:varchar(255);not null"`
	UID            string          `gorm:"column:uid;type:varchar(255);not null;index:idx_uid"`
	ChainName      string          `gorm:"column:chain_name;type:varchar(255);not null"`
	SourceTxID     string          `gorm:"column:source_tx_id;type:varchar(255);default:null;uniqueIndex:idx_source_tx_id"`
	DepositAddress string          `gorm:"column:deposit_address;type:varchar(255);not null"`
	Amount         decimal.Decimal `gorm:"column:amount;type:decimal(36,18);not null"`
	Fee            decimal.Decimal `gorm:"column:fee;type:decimal(36,18);not null"`
	Status         DepositStatus   `gorm:"column:status;type:tinyint(3);not null"`
	TxHash         string          `gorm:"column:tx_hash;type:varchar(255);not null;uniqueIndex:idx_tx_hash"`
	Time           time.Time       `gorm:"column:time;type:timestamp;not null"`
	BlockNumber    uint64          `gorm:"column:block_number;type:bigint(20);default:0"`
}

func (d *DepositRecord) TableName() string {
	return "deposit_records"
}

func (r *Repo) UpsertDepositRecord(ctx context.Context, record *DepositRecord) error {
	return r.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{"status", "tx_hash"}),
	}).Create(record).Error
}

func (r *Repo) GetDepositRecords(ctx context.Context, uid string, queryParams *queryparams.QueryParams) ([]*DepositRecord, int64, error) {
	queryParams.Add("uid", uid)
	var records []*DepositRecord
	total, err := r.Query(ctx, &records, queryParams, "uid")
	return records, total, err
}

func (r *Repo) GetDepositRecordBySourceTxID(ctx context.Context, sourceTxID string) (*DepositRecord, error) {
	var record DepositRecord
	err := r.WithContext(ctx).Where("source_tx_id = ?", sourceTxID).First(&record).Error
	return &record, err
}

func (r *Repo) GetDepositRecordByTxHash(ctx context.Context, txHash string) (*DepositRecord, error) {
	var record DepositRecord
	err := r.WithContext(ctx).Where("tx_hash = ?", txHash).First(&record).Error
	return &record, err
}

func (r *Repo) GetPendingDepositRecords(ctx context.Context, uid string) ([]*DepositRecord, error) {
	var records []*DepositRecord
	err := r.WithContext(ctx).Where("uid = ? and status = ?", uid, DepositStatusPending).Find(&records).Error
	return records, err
}

type UserDepositAddress struct {
	gorm.Model
	UID          string `gorm:"column:uid;type:varchar(255);not null;index:idx_uid"`
	Address      string `gorm:"column:address;type:varchar(255);not null;index:idx_address"`
	PermissionID uint64 `gorm:"column:permission_id;type:bigint(20);not null;index:idx_permission_id"`
	Remark       string `gorm:"column:remark;type:varchar(255);not null"`
}

func (u *UserDepositAddress) TableName() string {
	return "user_deposit_addresses"
}

func (r *Repo) CreateUserDepositAddress(ctx context.Context, address *UserDepositAddress) error {
	return r.DB.WithContext(ctx).Create(address).Error
}

func (r *Repo) GetUserDepositAddress(ctx context.Context, uid string, permissionID uint64) ([]UserDepositAddress, error) {
	var addresses []UserDepositAddress
	err := r.DB.WithContext(ctx).Where("uid = ? and permission_id = ?", uid, permissionID).Find(&addresses).Error
	return addresses, err
}

func (r *Repo) GetUIDByDepositAddress(ctx context.Context, address string) (string, error) {
	var uid string
	err := r.DB.WithContext(ctx).Model(&UserDepositAddress{}).Where("address = ?", address).Select("uid").Scan(&uid).Error
	return uid, err
}

func (r *Repo) GetUserDepositAddressByAddress(ctx context.Context, address string) (*UserDepositAddress, error) {
	var userDepositAddress UserDepositAddress
	err := r.DB.WithContext(ctx).Where("address = ?", address).First(&userDepositAddress).Error
	return &userDepositAddress, err
}

func (r *Repo) GetDepositMaxBlockNumber(ctx context.Context) (uint64, error) {
	var blockNumber *uint64
	err := r.WithContext(ctx).Model(&DepositRecord{}).Select("COALESCE(MAX(block_number), 0)").Scan(&blockNumber).Error
	if err != nil {
		return 0, err
	}
	if blockNumber == nil {
		return 0, nil
	}
	return *blockNumber, nil
}

type WithdrawStatus uint8

const (
	WithdrawStatusPending WithdrawStatus = iota
	WithdrawStatusSuccess
	WithdrawStatusFailed
)

type WithdrawRecord struct {
	gorm.Model
	UID         string          `gorm:"column:uid;type:varchar(255);not null"`
	Symbol      string          `gorm:"column:symbol;type:varchar(255);not null"`
	ChainName   string          `gorm:"column:chain_name;type:varchar(255);not null"`
	Amount      decimal.Decimal `gorm:"column:amount;type:decimal(36,18);not null"`
	Fee         decimal.Decimal `gorm:"column:fee;type:decimal(36,18);not null"`
	BridgeFee   decimal.Decimal `gorm:"column:bridge_fee;type:decimal(36,18);default:0"`
	Status      WithdrawStatus  `gorm:"column:status;type:tinyint(3);not null"`
	SendTxID    string          `gorm:"column:send_tx_id;type:varchar(255);default:null;uniqueIndex:idx_send_tx_id"`
	TxHash      string          `gorm:"column:tx_hash;type:varchar(255);not null;uniqueIndex:idx_tx_hash"`
	WithdrawAt  time.Time       `gorm:"column:withdraw_at;type:timestamp;default:null"`
	CompletedAt time.Time       `gorm:"column:completed_at;type:timestamp;default:null"`
	BlockNumber uint64          `gorm:"column:block_number;type:bigint(20);default:0"`
	Recipient   string          `gorm:"column:recipient;type:varchar(255);default:null"`
}

func (u *WithdrawRecord) TableName() string {
	return "withdraw_records"
}

func (r *Repo) CreateWithdrawRecord(ctx context.Context, record *WithdrawRecord) error {
	return r.DB.WithContext(ctx).Create(record).Error
}

func (r *Repo) UpdateWithdrawRecord(ctx context.Context, record *WithdrawRecord) error {
	return r.DB.WithContext(ctx).Model(&WithdrawRecord{}).Where("id = ?", record.ID).Updates(record).Error
}

func (r *Repo) GetWithdrawRecordByTxHash(ctx context.Context, txHash string) (*WithdrawRecord, error) {
	var record WithdrawRecord
	err := r.DB.WithContext(ctx).Where("tx_hash = ?", txHash).First(&record).Error
	return &record, err
}

func (r *Repo) GetWithdrawRecords(ctx context.Context, uid string, queryParams *queryparams.QueryParams) ([]*WithdrawRecord, int64, error) {
	queryParams.Add("uid", uid)
	var records []*WithdrawRecord
	total, err := r.Query(ctx, &records, queryParams, "uid")
	return records, total, err
}

func (r *Repo) GetWithdrawMaxBlockNumber(ctx context.Context) (uint64, error) {
	var blockNumber *uint64
	err := r.WithContext(ctx).Model(&WithdrawRecord{}).Select("COALESCE(MAX(block_number), 0)").Scan(&blockNumber).Error
	if err != nil {
		return 0, err
	}
	if blockNumber == nil {
		return 0, nil
	}
	return *blockNumber, nil
}

func (r *Repo) GetPendingWithdrawRecords(ctx context.Context, uid string) ([]*WithdrawRecord, error) {
	var records []*WithdrawRecord
	err := r.WithContext(ctx).Where("uid = ? and status = ?", uid, WithdrawStatusPending).Find(&records).Error
	return records, err
}

func (r *Repo) GetAllPendingDepositRecords(ctx context.Context) ([]*DepositRecord, error) {
	var records []*DepositRecord
	err := r.WithContext(ctx).
		Select("id, uid, symbol, amount, chain_name, deposit_address, time, tx_hash").
		Where("status = ?", DepositStatusPending).
		Find(&records).Error
	return records, err
}

func (r *Repo) GetAllPendingWithdrawRecords(ctx context.Context) ([]*WithdrawRecord, error) {
	var records []*WithdrawRecord
	err := r.WithContext(ctx).
		Select("id, uid, symbol, amount, chain_name, fee, bridge_fee, recipient, withdraw_at, tx_hash").
		Where("status = ?", WithdrawStatusPending).
		Find(&records).Error
	return records, err
}

type TransactionsRecord struct {
	ID             uint      ``
	DepositAt      time.Time `json:"deposit_at"`
	WithdrawAt     time.Time `json:"withdraw_at"`
	Symbol         string    `json:"symbol"`
	CoinName       string    `json:"coin_name"`
	EVMAddress     string    `json:"evm_address"`
	UID            string    `json:"uid"`
	Fee            float64   `json:"fee"`
	DepositChain   string    `json:"deposit_chain"`
	WithdrawChain  string    `json:"withdraw_chain"`
	DepositAmount  float64   `json:"deposit_amount"`
	WithdrawAmount float64   `json:"withdraw_amount"`
	TxHash         string    `json:"tx_hash"`
}

func (r *Repo) QueryTransactionsRecord(ctx context.Context, params *queryparams.QueryParams) ([]*TransactionsRecord, int64, error) {
	var records []*TransactionsRecord

	var whereConditions []string
	var queryParams []interface{}

	if symbol, ok := params.CustomQuery["symbol"]; ok {
		whereConditions = append(whereConditions, "symbol = ?")
		queryParams = append(queryParams, symbol[0])
	}
	if chainName, ok := params.CustomQuery["chain_name"]; ok {
		whereConditions = append(whereConditions, "chain_name = ?")
		queryParams = append(queryParams, chainName[0])
	}
	if uid, ok := params.CustomQuery["uid"]; ok {
		whereConditions = append(whereConditions, "uid = ?")
		queryParams = append(queryParams, uid[0])
	}
	if sendTxId, ok := params.CustomQuery["tx_hash"]; ok {
		whereConditions = append(whereConditions, "tx_hash = ?")
		queryParams = append(queryParams, sendTxId[0])
	}
	if startTime, ok := params.CustomQuery["start_time"]; ok {
		whereConditions = append(whereConditions, "created_at > ?")
		queryParams = append(queryParams, startTime[0])
	}
	if endTime, ok := params.CustomQuery["end_time"]; ok {
		whereConditions = append(whereConditions, "created_at <= ?")
		queryParams = append(queryParams, endTime[0])
	}

	whereSQL := ""
	if len(whereConditions) > 0 {
		whereSQL = "WHERE " + strings.Join(whereConditions, " AND ")
	}

	querySQL := fmt.Sprintf(`
        SELECT id, created_at AS deposit_at, symbol, uid, fee, chain_name AS deposit_chain, amount AS deposit_amount, tx_hash FROM deposit_records %s
        UNION ALL
        SELECT id, created_at AS withdraw_at, symbol, uid, fee, chain_name AS withdraw_chain, amount AS withdraw_amount, tx_hash FROM withdraw_records %s
        ORDER BY id DESC LIMIT ? OFFSET ?
    `, whereSQL, whereSQL)

	queryParams = append(queryParams, queryParams...)
	queryParams = append(queryParams, params.Limit, params.Offset)

	fmt.Println(querySQL)
	fmt.Println(queryParams)
	err := r.Raw(querySQL, queryParams...).Scan(&records).Error
	if err != nil {
		return nil, 0, err
	}

	var total int64
	countSQL := fmt.Sprintf(`
        SELECT COUNT(*) FROM (
            SELECT id FROM deposit_records %s
            UNION ALL
            SELECT id FROM withdraw_records %s
        ) AS total_count
    `, whereSQL, whereSQL)

	err = r.Raw(countSQL, queryParams[:len(queryParams)-2]...).Scan(&total).Error
	if err != nil {
		return nil, 0, err
	}

	return records, total, nil
}
