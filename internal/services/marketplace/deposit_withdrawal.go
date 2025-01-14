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
	return nil, 0, nil
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

	remark := fmt.Sprintf("new-%s-%s", uid, req.PublicKey)
	depositAddress, err := s.repo.GetUserDepositAddress(ctx, uid, req.Symbol, req.ChainName)
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

	token, err := s.repo.GetToken(ctx, req.Symbol, req.ChainName)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return entity.RespFirstDeposit{}, errors.New("token not found")
		}
		return entity.RespFirstDeposit{}, err
	}

	bridgeClient := eos.NewBridgeClient(
		s.eosCfg.NodeURL,
		s.eosCfg.Bridge,
		s.eosCfg.Actor,
		s.eosCfg.ActorPrivateKey,
	)

	resp, err := bridgeClient.MappingAddress(ctx, eos.MappingAddrRequest{
		PermissionID:     token.PermissionID,
		RecipientAddress: s.eosCfg.VaultEVMAddress,
		Remark:           remark,
	})
	if err != nil {
		return entity.RespFirstDeposit{}, err
	}
	log.Printf("mapping new address txid: %v", resp.TransactionID)

	time.Sleep(1 * time.Second)

	newDepositAddress, err := bridgeClient.GetDepositAddress(ctx, eos.RequestDepositAddress{
		PermissionID: token.PermissionID,
		Remark:       remark,
		Recipient:    s.eosCfg.VaultEVMAddress,
	})
	if err != nil {
		log.Printf("get deposit address from bridge error: %v", err)
		return entity.RespFirstDeposit{}, err
	}
	log.Printf("get first deposit address: %s", newDepositAddress)
	err = s.repo.CreateUserDepositAddress(ctx, &db.UserDepositAddress{
		UID:       uid,
		Address:   newDepositAddress,
		Symbol:    req.Symbol,
		ChainName: req.ChainName,
		Remark:    remark,
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

	depositAddress, err := s.repo.GetUserDepositAddress(ctx, uid, req.Symbol, req.ChainName)
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
	
	token, err := s.repo.GetToken(ctx, req.Symbol, req.ChainName)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return entity.RespDeposit{}, errors.New("token not found")
		}
		return entity.RespDeposit{}, err
	}

	bridgeClient := eos.NewBridgeClient(
		s.eosCfg.NodeURL,
		s.eosCfg.Bridge,
		s.eosCfg.Actor,
		s.eosCfg.ActorPrivateKey,
	)

	resp, err := bridgeClient.MappingAddress(ctx, eos.MappingAddrRequest{
		PermissionID:     token.PermissionID,
		RecipientAddress: s.eosCfg.VaultEVMAddress,
		Remark:          remark,
	})
	if err != nil {
		return entity.RespDeposit{}, err
	}
	log.Printf("mapping new address txid: %v", resp.TransactionID)

	time.Sleep(1 * time.Second)

	newDepositAddress, err := bridgeClient.GetDepositAddress(ctx, eos.RequestDepositAddress{
		PermissionID: token.PermissionID,
		Remark:       remark,
		Recipient:    s.eosCfg.VaultEVMAddress,
	})
	if err != nil {
		log.Printf("get deposit address from bridge error: %v", err)
		return entity.RespDeposit{}, err
	}
	log.Printf("new deposit address: %s", newDepositAddress)
	err = s.repo.CreateUserDepositAddress(ctx, &db.UserDepositAddress{
		UID:       uid,
		Address:   newDepositAddress,
		Symbol:    req.Symbol,
		ChainName: req.ChainName,
		Remark:    remark,
	})
	if err != nil {
		return entity.RespDeposit{}, err
	}
	return entity.RespDeposit{
		Address: newDepositAddress,
	}, nil
}

