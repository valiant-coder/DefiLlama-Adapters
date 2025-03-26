package db

import (
	"context"
	"exapp-go/pkg/queryparams"
	"fmt"
	"math/rand"
	"time"

	"gorm.io/gorm"
)

func init() {
	addMigrateFunc(func(r *Repo) error {
		return r.AutoMigrate(
			&User{},
			&UserCredential{},
		)
	})
}

type LoginMethod string

const (
	LoginMethodGoogle   LoginMethod = "google"
	LoginMethodApple    LoginMethod = "apple"
	LoginMethodTelegram LoginMethod = "telegram"
	LoginMethodEVM      LoginMethod = "evm"
)

type User struct {
	gorm.Model
	Username    string      `gorm:"column:username;type:varchar(255);default:null;index:idx_username"`
	UID         string      `gorm:"column:uid;type:varchar(255);not null;uniqueIndex:idx_uid"`
	LoginMethod LoginMethod `gorm:"column:login_method;type:varchar(255);default:null;uniqueIndex:idx_login_method_oauth_id"`
	Avatar      string      `gorm:"column:avatar;type:varchar(255);default:null"`
	OauthID     string      `gorm:"column:oauth_id;type:varchar(255);default:null;uniqueIndex:idx_login_method_oauth_id"`
	Email       string      `gorm:"column:email;type:varchar(255);default:null"`

	// for evm user
	EVMAddress string `gorm:"column:evm_address;type:varchar(255);default:null;index:idx_evm_address"`
	EOSAccount string `gorm:"column:eos_account;type:varchar(255);default:null;index:idx_eos_account_permission"`
	Permission string `gorm:"column:permission;type:varchar(255);default:null;index:idx_eos_account_permission"`

	BlockNumber uint64 `gorm:"column:block_number;default:0;type:bigint(20)"`
}

func (User) TableName() string {
	return "users"
}

func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	randomNum := r.Intn(90000000) + 10000000
	u.UID = fmt.Sprintf("%d", randomNum)
	return
}

func (r *Repo) UpsertUser(ctx context.Context, user *User) error {
	var existingUser User
	err := r.DB.WithContext(ctx).Where(
		"login_method = ? AND oauth_id = ?",
		user.LoginMethod,
		user.OauthID,
	).First(&existingUser).Error
	if err == nil {
		existingUser.Avatar = user.Avatar
		existingUser.Username = user.Username
		existingUser.Email = user.Email
		existingUser.BlockNumber = user.BlockNumber
		*user = existingUser
		return r.DB.WithContext(ctx).Save(user).Error
	}
	return r.DB.WithContext(ctx).Create(user).Error
}

func (r *Repo) UpdateUser(ctx context.Context, user *User) error {
	return r.DB.WithContext(ctx).Model(&User{}).Where("id =?", user.ID).Updates(user).Error
}

func (r *Repo) IsUserExist(ctx context.Context, uid string) (bool, error) {
	var user User
	result := r.DB.WithContext(ctx).Where("uid = ?", uid).First(&user)
	return result.RowsAffected > 0, result.Error
}

func (r *Repo) GetUser(ctx context.Context, uid string) (*User, error) {
	var user User
	err := r.DB.WithContext(ctx).Where("uid = ?", uid).First(&user).Error
	return &user, err
}

func (r *Repo) GetUserByEVMAddress(ctx context.Context, evmAddress string) (*User, error) {
	var user User
	result := r.DB.WithContext(ctx).Where("evm_address =?", evmAddress).First(&user)
	return &user, result.Error
}

func (r *Repo) GetTotalUserCount(ctx context.Context) (int64, error) {
	var totalUserCount int64
	result := r.DB.WithContext(ctx).Model(&User{}).Count(&totalUserCount).Error
	if result != nil {
		return 0, result
	}
	return totalUserCount, nil
}

func (r *Repo) GetEOSAccountAndPermissionByUID(ctx context.Context, uid string) (string, string, error) {
	var user User
	result := r.DB.WithContext(ctx).Where("uid = ?", uid).First(&user)
	if result.Error != nil {
		return "", "", result.Error
	}
	if user.EVMAddress != "" {
		return user.EOSAccount, user.Permission, nil
	}
	credentials, err := r.GetUserCredentials(ctx, uid)
	if err != nil {
		return "", "", err
	}
	for _, c := range credentials {
		if c.EOSAccount != "" {
			return c.EOSAccount, "active", nil
		}
	}
	return "", "", nil
}

func (r *Repo) GetUserMaxBlockNumber(ctx context.Context) (uint64, error) {
	var blockNumber *uint64
	err := r.WithContext(ctx).Model(&User{}).Select("COALESCE(MAX(block_number), 0)").Scan(&blockNumber).Error
	if err != nil {
		return 0, err
	}
	if blockNumber == nil {
		return 0, nil
	}
	return *blockNumber, nil
}

type UserCredential struct {
	gorm.Model
	UID            string    `gorm:"column:uid;type:varchar(255);not null;index:idx_uid"`
	CredentialID   string    `gorm:"column:credential_id;type:varchar(255);uniqueIndex:idx_credential_id"`
	PublicKey      string    `gorm:"column:public_key;type:varchar(255);not null;index:idx_public_key"`
	Name           string    `gorm:"column:name;type:varchar(255);not null"`
	LastUsedAt     time.Time `gorm:"column:last_used_at;default:null;type:timestamp"`
	LastUsedIP     string    `gorm:"column:last_used_ip;type:varchar(255)"`
	Synced         bool      `gorm:"column:synced;type:tinyint(1);not null;default:0"`
	EOSAccount     string    `gorm:"column:eos_account;type:varchar(255);index:idx_eos_account"`
	EOSPermissions string    `gorm:"column:eos_permissions;type:varchar(512)"`
	DeviceID       string    `gorm:"column:device_id;default:null;type:varchar(255);index:idx_device_id"`
	BlockNumber    uint64    `gorm:"column:block_number;default:0;type:bigint(20)"`
	AAGuid         string    `gorm:"column:aaguid;type:varchar(255);default:null"`
	Storage        string    `gorm:"column:storage;type:varchar(255);default:null"`
}

func (UserCredential) TableName() string {
	return "user_credentials"
}

func (r *Repo) CreateCredentialIfNotExist(ctx context.Context, credential *UserCredential) error {
	var existingCredential UserCredential
	result := r.DB.WithContext(ctx).Where("credential_id = ?", credential.CredentialID).First(&existingCredential)
	if result.Error == gorm.ErrRecordNotFound {
		return r.DB.WithContext(ctx).Create(credential).Error
	}
	return nil
}

func (r *Repo) GetUserCredentials(ctx context.Context, uid string) ([]*UserCredential, error) {
	var credentials []*UserCredential
	result := r.DB.WithContext(ctx).Where("uid = ?", uid).Find(&credentials)
	return credentials, result.Error
}

func (r *Repo) GetUserCredentialByPubkey(ctx context.Context, pubkey string) (*UserCredential, error) {
	var credential UserCredential
	result := r.DB.WithContext(ctx).Where("public_key = ?", pubkey).First(&credential)
	return &credential, result.Error
}

func (r *Repo) GetUserCredentialsByEOSAccount(ctx context.Context, eosAccount string) ([]*UserCredential, error) {
	var credentials []*UserCredential
	err := r.DB.WithContext(ctx).Where("eos_account = ?", eosAccount).Find(&credentials).Error
	if err != nil {
		return nil, err
	}
	return credentials, nil
}

func (r *Repo) GetEosAccountByUID(ctx context.Context, uid string) (string, error) {
	var credential UserCredential
	err := r.DB.WithContext(ctx).Where("uid = ? and eos_account != ''", uid).First(&credential).Error
	if err != nil {
		return "", err
	}
	return credential.EOSAccount, nil
}

func (r *Repo) GetUserCredentialsByKeys(ctx context.Context, keys []string) ([]*UserCredential, error) {
	var credentials []*UserCredential
	err := r.DB.WithContext(ctx).Where("public_key IN (?)", keys).Find(&credentials).Error
	if err != nil {
		return nil, err
	}
	return credentials, nil
}

func (r *Repo) UpdateUserCredential(ctx context.Context, credential *UserCredential) error {
	return r.DB.WithContext(ctx).Model(&UserCredential{}).Where("id = ?", credential.ID).Updates(credential).Error
}

func (r *Repo) DeleteUserCredential(ctx context.Context, credential *UserCredential) error {
	return r.DB.WithContext(ctx).Where("id = ?", credential.ID).Delete(&UserCredential{}).Error
}

func (r *Repo) GetUIDByEOSAccount(ctx context.Context, eosAccount string) (string, error) {
	var credential UserCredential
	result := r.DB.WithContext(ctx).Where("eos_account = ?", eosAccount).First(&credential)
	return credential.UID, result.Error
}

func (r *Repo) GetUserCredentialMaxBlockNumber(ctx context.Context) (uint64, error) {
	var blockNumber *uint64
	err := r.WithContext(ctx).Model(&UserCredential{}).Select("COALESCE(MAX(block_number), 0)").Scan(&blockNumber).Error
	if err != nil {
		return 0, err
	}
	if blockNumber == nil {
		return 0, nil
	}
	return *blockNumber, nil
}

type EOSAccountInfo struct {
	UID        string `gorm:"column:uid;"`
	EOSAccount string `gorm:"column:eos_account;"`
}

func (r *Repo) GetAllEOSAccounts(ctx context.Context) ([]*EOSAccountInfo, error) {
	var accounts []*EOSAccountInfo
	err := r.DB.WithContext(ctx).
		Model(&UserCredential{}).
		Select("DISTINCT uid, eos_account").
		Where("eos_account != ''").
		Find(&accounts).Error
	return accounts, err
}

func (r *Repo) GetUsersByUIDs(ctx context.Context, uids []string) (map[string]User, error) {
	var users []User
	err := r.DB.WithContext(ctx).Where("uid IN ?", uids).Find(&users).Error
	if err != nil {
		return nil, err
	}

	userMap := make(map[string]User)
	for _, user := range users {
		userMap[user.UID] = user
	}
	return userMap, nil
}

type UserList struct {
	ID             uint        `gorm:"column:id"`
	Username       string      `gorm:"column:username"`
	UID            string      `gorm:"column:uid"`
	LoginMethod    LoginMethod `gorm:"column:login_method"`
	CreatedAt      time.Time   `gorm:"column:created_at"`
	PasskeyCount   int
	LastUsedAt     time.Time
	FirstDepositAt time.Time
	LastDepositAt  time.Time
	LastWithdrawAt time.Time
}

func (r *Repo) QueryUserList(ctx context.Context, params *queryparams.QueryParams) ([]*UserList, int64, error) {
	var users []*UserList

	oldestDepositRecords := r.DB.Table("deposit_records").
		Select("uid, MIN(time) AS first_deposit_at").
		Where("deleted_at IS NULL").
		Group("uid")

	newestDepositRecords := r.DB.Table("deposit_records").
		Select("uid, MIN(time) AS last_deposit_at").
		Where("deleted_at IS NULL").
		Group("uid")

	newestWithdrawRecords := r.DB.Table("withdraw_records").
		Select("uid, MAX(withdraw_at) AS last_withdraw_at").
		Where("deleted_at IS NULL").
		Group("uid")

	passkeyCount := r.DB.Table("user_credentials").
		Select("uid, COUNT(*) AS passkey_count").
		Where("deleted_at IS NULL").
		Group("uid")

	lastUsedAt := r.DB.Table("user_credentials").
		Select("uid, MAX(last_used_at) AS last_used_at").
		Where("deleted_at IS NULL").
		Group("uid")

	queryable := []string{"login_method"}
	tx := r.DB.Table("users").Limit(params.Limit).Offset(params.Offset).Order(params.Order)

	plains := params.Query.Plains(queryable...)
	if len(plains) > 0 {
		tx = tx.Where(plains[0], plains[1:]...)
	}
	if uid, ok := params.CustomQuery["uid"]; ok {
		tx = tx.Where("users.uid = ?", uid[0])
	}
	if username, ok := params.CustomQuery["username"]; ok {
		tx = tx.Where("username = ?", username[0])
	}
	if startTime, ok := params.CustomQuery["start_time"]; ok {
		tx = tx.Where("created_at >= ?", startTime[0])
	}
	if endTime, ok := params.CustomQuery["end_time"]; ok {
		tx = tx.Where("created_at <= ?", endTime[0])
	}
	err := tx.Select("users.*,"+
		"newestWithdraw.last_withdraw_at,"+"oldestDeposit.first_deposit_at,"+
		"newestDeposit.last_deposit_at,"+"pc.passkey_count, lu.last_used_at").
		Joins("LEFT JOIN (?) AS oldestDeposit ON users.uid = oldestDeposit.uid", oldestDepositRecords).
		Joins("LEFT JOIN (?) AS newestDeposit ON users.uid = newestDeposit.uid", newestDepositRecords).
		Joins("LEFT JOIN (?) AS newestWithdraw ON users.uid = newestWithdraw.uid", newestWithdrawRecords).
		Joins("LEFT JOIN (?) AS pc ON users.uid = pc.uid", passkeyCount).
		Joins("LEFT JOIN (?) AS lu ON users.uid = lu.uid", lastUsedAt).
		Find(&users).Error
	if err != nil {
		return nil, 0, err
	}

	var total int64
	err = r.DB.Table("users").Limit(-1).Offset(-1).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

type UsersStatis struct {
	Period string `json:"period"`
	Count  int64  `json:"count"`
}

func (r *Repo) GetStatisAddUserCount(ctx context.Context, dimension string, amount int) ([]*UsersStatis, int64, error) {

	var format, param string
	switch dimension {
	case "day":
		format = "DATE_FORMAT(created_at, '%Y-%m-%d')"
		param = fmt.Sprintf("%d DAY", amount)
	case "week":
		format = "CONCAT(YEAR(created_at), '-W', LPAD(WEEK(created_at, 3), 2, '0'))"
		param = fmt.Sprintf("%d WEEK", amount)
	case "month":
		format = "DATE_FORMAT(created_at, '%Y-%m')"
		param = fmt.Sprintf("%d MONTH", amount)
	default:
		return nil, 0, fmt.Errorf("time dimension is invalid")
	}

	sql := fmt.Sprintf(`SELECT %s AS period, COUNT(*) AS count
		FROM users
		WHERE created_at >= CURDATE() - INTERVAL %s
		GROUP BY period
		ORDER BY period;`, format, param)
	var data []*UsersStatis
	err := r.DB.Raw(sql).Scan(&data).Error
	if err != nil {
		return nil, 0, err
	}

	var total int64
	err = r.DB.Table("users").Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	data = fillMissingDates(data, dimension, amount)
	return data, total, nil
}

func (r *Repo) GetStatisAddPasskeyCount(ctx context.Context, dimension string, amount int) ([]*UsersStatis, int64, error) {

	var format, param string
	switch dimension {
	case "day":
		format = "DATE_FORMAT(users.created_at, '%Y-%m-%d')"
		param = fmt.Sprintf("%d DAY", amount)
	case "week":
		format = "CONCAT(YEAR(users.created_at), '-W', LPAD(WEEK(users.created_at, 3), 2, '0'))"
		param = fmt.Sprintf("%d WEEK", amount)
	case "month":
		format = "DATE_FORMAT(users.created_at, '%Y-%m')"
		param = fmt.Sprintf("%d MONTH", amount)
	default:
		return nil, 0, fmt.Errorf("time dimension is invalid")
	}

	sql := fmt.Sprintf(`SELECT %s AS period, COUNT(DISTINCT users.uid) AS count
		FROM users
		INNER JOIN user_credentials ON users.uid = user_credentials.uid
		WHERE users.created_at >= DATE_SUB(CURDATE(), INTERVAL %s) 
		GROUP BY period
		ORDER BY period;`, format, param)
	var data []*UsersStatis
	err := r.DB.Raw(sql).Scan(&data).Error
	if err != nil {
		return nil, 0, err
	}

	var total int64
	err = r.DB.Table("users").
		Select("COUNT(DISTINCT users.uid)").
		Joins("LEFT JOIN user_credentials ON users.uid = user_credentials.uid").
		Where("user_credentials.uid IS NOT NULL").
		Scan(&total).Error
	if err != nil {
		return nil, 0, err
	}

	data = fillMissingDates(data, dimension, amount)
	return data, total, nil
}

// Data when the number of fills is 0
func fillMissingDates(rawStatis []*UsersStatis, dimension string, amount int) []*UsersStatis {
	var layout string
	var step time.Duration
	statis := make(map[string]int64)

	switch dimension {
	case "day":
		layout = "2006-01-02"
		step = 24 * time.Hour
	case "week":
		layout = "2006-W01"
		step = 7 * 24 * time.Hour
	case "month":
		layout = "2006-01"
		step = 0
	default:
		return nil
	}

	for _, entry := range rawStatis {
		statis[entry.Period] = entry.Count
	}

	now := time.Now()
	var startDate time.Time

	if dimension == "day" {
		startDate = now.AddDate(0, 0, -amount+1)
	} else if dimension == "week" {
		startDate = now.AddDate(0, 0, -7*(amount-1))
		for startDate.Weekday() != time.Monday {
			startDate = startDate.AddDate(0, 0, -1)
		}
	} else if dimension == "month" {
		startDate = now.AddDate(0, -amount+1, 0)
		startDate = time.Date(startDate.Year(), startDate.Month(), 1, 0, 0, 0, 0, time.Local)
	}

	var filledStats []*UsersStatis
	for i := 0; i < amount; i++ {
		var dateStr string

		if dimension == "week" {
			year, week := startDate.ISOWeek()
			dateStr = fmt.Sprintf("%d-W%02d", year, week)
		} else {
			dateStr = startDate.Format(layout)
		}

		count := statis[dateStr]
		filledStats = append(filledStats, &UsersStatis{
			Period: dateStr,
			Count:  count,
		})

		if dimension == "month" {
			startDate = startDate.AddDate(0, 1, 0)
		} else {
			startDate = startDate.Add(step)
		}
	}

	return filledStats
}
