package marketplace

import (
	"context"
	"errors"
	"exapp-go/config"
	"exapp-go/internal/db/db"
	"exapp-go/internal/entity"
	"exapp-go/pkg/eos"
	"exapp-go/pkg/queryparams"
	"fmt"
	"log"
	"time"

	"gorm.io/gorm"
)

type DepositWithdrawalService struct {
	repo   *db.Repo
	eosCfg config.EosConfig
}

func NewDepositWithdrawalService() *DepositWithdrawalService {
	return &DepositWithdrawalService{
		repo:   db.New(),
		eosCfg: config.Conf().Eos,
	}
}

func (s *DepositWithdrawalService) GetDepositRecords(ctx context.Context, uid string, queryParams *queryparams.QueryParams) ([]entity.RespDepositRecord, int64, error) {
	records, total, err := s.repo.GetDepositRecords(ctx, uid, queryParams)
	if err != nil {
		return nil, 0, err
	}
	var result []entity.RespDepositRecord
	for _, record := range records {
		result = append(result, entity.FormatDepositRecord(record))
	}
	return result, total, nil
}

func (s *DepositWithdrawalService) FirstDeposit(ctx context.Context, uid string, req *entity.ReqFirstDeposit) (entity.RespFirstDeposit, error) {
	passkeys, err := s.repo.GetUserCredentials(ctx, uid)
	if err != nil {
		return entity.RespFirstDeposit{}, err
	}
	if len(passkeys) == 0 {
		return entity.RespFirstDeposit{}, errors.New("no passkey found")
	}
	var hasPasskey bool
	for _, passkey := range passkeys {
		if passkey.EOSAccount != "" {
			return entity.RespFirstDeposit{}, errors.New("had first deposit,please use deposit")
		}
		if passkey.PublicKey == req.PublicKey {
			hasPasskey = true
			break
		}
	}
	if !hasPasskey {
		return entity.RespFirstDeposit{}, errors.New("no found passkey")
	}

	token, err := s.repo.GetToken(ctx, req.Symbol, req.ChainName)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return entity.RespFirstDeposit{}, errors.New("token not found")
		}
		return entity.RespFirstDeposit{}, err
	}
	remark := fmt.Sprintf("new-%s-%s", uid, req.PublicKey)
	depositAddress, err := s.repo.GetUserDepositAddress(ctx, uid, token.PermissionID)
	if err != nil {
		return entity.RespFirstDeposit{}, err
	}
	if len(depositAddress) > 0 {
		for _, address := range depositAddress {
			if address.Remark == remark {
				return entity.RespFirstDeposit{
					Address: address.Address,
				}, nil
			}
		}
	}

	bridgeClient := eos.NewBridgeClient(
		s.eosCfg.NodeURL,
		s.eosCfg.Exsat.BridgeContract,
		s.eosCfg.Exapp.Actor,
		s.eosCfg.Exapp.ActorPrivateKey,
	)

	resp, err := bridgeClient.MappingAddress(ctx, eos.MappingAddrRequest{
		PermissionID:     token.PermissionID,
		RecipientAddress: s.eosCfg.Exapp.VaultEVMAddress,
		Remark:           remark,
	})
	if err != nil {
		return entity.RespFirstDeposit{}, err
	}
	log.Printf("mapping new address txid: %v", resp.TransactionID)

	time.Sleep(500 * time.Millisecond)

	newDepositAddress, err := bridgeClient.GetDepositAddress(ctx, eos.RequestDepositAddress{
		PermissionID: token.PermissionID,
		Remark:       remark,
		Recipient:    s.eosCfg.Exapp.VaultEVMAddress,
	})
	if err != nil {
		log.Printf("get deposit address from bridge error: %v", err)
		return entity.RespFirstDeposit{}, err
	}
	log.Printf("get first deposit address: %s", newDepositAddress)
	err = s.repo.CreateUserDepositAddress(ctx, &db.UserDepositAddress{
		UID:          uid,
		Address:      newDepositAddress,
		PermissionID: token.PermissionID,
		Remark:       remark,
	})
	if err != nil {
		return entity.RespFirstDeposit{}, err
	}
	return entity.RespFirstDeposit{
		Address: newDepositAddress,
	}, nil

}

func (s *DepositWithdrawalService) Deposit(ctx context.Context, uid string, req *entity.ReqDeposit) (entity.RespDeposit, error) {
	passkeys, err := s.repo.GetUserCredentials(ctx, uid)
	if err != nil {
		return entity.RespDeposit{}, err
	}
	if len(passkeys) == 0 {
		return entity.RespDeposit{}, errors.New("no passkey found")
	}
	var eosAccount string
	for _, passkey := range passkeys {
		if passkey.EOSAccount != "" {
			eosAccount = passkey.EOSAccount
			break
		}
	}
	if eosAccount == "" {
		return entity.RespDeposit{}, errors.New("no eos account found")
	}

	token, err := s.repo.GetToken(ctx, req.Symbol, req.ChainName)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return entity.RespDeposit{}, errors.New("token not found")
		}
		return entity.RespDeposit{}, err
	}

	depositAddress, err := s.repo.GetUserDepositAddress(ctx, uid, token.PermissionID)
	if err != nil {
		return entity.RespDeposit{}, err
	}

	remark := fmt.Sprintf("deposit-%s", eosAccount)
	if len(depositAddress) > 0 {
		for _, address := range depositAddress {
			if address.Remark == remark {
				return entity.RespDeposit{
					Address: address.Address,
				}, nil
			}
		}
	}

	bridgeClient := eos.NewBridgeClient(
		s.eosCfg.NodeURL,
		s.eosCfg.Exsat.BridgeContract,
		s.eosCfg.Exapp.Actor,
		s.eosCfg.Exapp.ActorPrivateKey,
	)

	resp, err := bridgeClient.MappingAddress(ctx, eos.MappingAddrRequest{
		PermissionID:     token.PermissionID,
		RecipientAddress: s.eosCfg.Exapp.VaultEVMAddress,
		Remark:           remark,
	})
	if err != nil {
		return entity.RespDeposit{}, err
	}
	log.Printf("mapping new address txid: %v", resp.TransactionID)

	time.Sleep(500 * time.Millisecond)

	newDepositAddress, err := bridgeClient.GetDepositAddress(ctx, eos.RequestDepositAddress{
		PermissionID: token.PermissionID,
		Remark:       remark,
		Recipient:    s.eosCfg.Exapp.VaultEVMAddress,
	})
	if err != nil {
		log.Printf("get deposit address from bridge error: %v", err)
		return entity.RespDeposit{}, err
	}
	log.Printf("new deposit address: %s", newDepositAddress)
	err = s.repo.CreateUserDepositAddress(ctx, &db.UserDepositAddress{
		UID:          uid,
		Address:      newDepositAddress,
		PermissionID: token.PermissionID,
		Remark:       remark,
	})
	if err != nil {
		return entity.RespDeposit{}, err
	}
	return entity.RespDeposit{
		Address: newDepositAddress,
	}, nil
}

func (s *DepositWithdrawalService) GetWithdrawalRecords(ctx context.Context, uid string, queryParams *queryparams.QueryParams) ([]entity.RespWithdrawRecord, int64, error) {
	records, total, err := s.repo.GetWithdrawRecords(ctx, uid, queryParams)
	if err != nil {
		return nil, 0, err
	}
	var result []entity.RespWithdrawRecord
	for _, record := range records {
		result = append(result, entity.FormatWithdrawRecord(record))
	}
	return result, total, nil
}
