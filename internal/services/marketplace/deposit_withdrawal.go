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

func (s *DepositWithdrawalService) pollForDepositAddress(ctx context.Context, bridgeClient *eos.BridgeClient, req eos.RequestDepositAddress) (string, error) {
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		address, err := bridgeClient.GetDepositAddress(ctx, req)
		if err == nil && address != "" {
			return address, nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return "", errors.New("timeout waiting for deposit address")
}

func (s *DepositWithdrawalService) pollForBTCAddress(ctx context.Context, bridgeClient *eos.BTCBridgeClient, req eos.RequestBTCDepositAddress) (string, error) {
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		address, err := bridgeClient.GetDepositAddress(ctx, req)
		if err == nil && address != "" {
			return address, nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return "", errors.New("timeout waiting for deposit address")
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

	token, err := s.repo.GetToken(ctx, req.Symbol)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return entity.RespFirstDeposit{}, errors.New("token not found")
		}
		return entity.RespFirstDeposit{}, err
	}
	var targetChain db.ChainInfo
	for _, chain := range token.Chains {
		if chain.ChainName == req.ChainName {
			targetChain = chain
			break
		}
	}
	if targetChain.ChainName == "" {
		return entity.RespFirstDeposit{}, errors.New("chain not found")
	}
	remark := fmt.Sprintf("new-%s-%s", uid, req.PublicKey)
	depositAddress, err := s.repo.GetUserDepositAddress(ctx, uid, targetChain.PermissionID)
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


	var newDepositAddress string
	if req.Symbol == "BTC" && req.ChainName != "eos" && req.ChainName != "exsat" {
		btcBridgeClient := eos.NewBTCBridgeClient(
			s.eosCfg.NodeURL,
			s.eosCfg.Exsat.BTCBridgeContract,
			s.eosCfg.Exapp.Actor,
			s.eosCfg.Exapp.ActorPrivateKey,
		)

		resp, err := btcBridgeClient.MappingAddress(ctx, eos.BTCMappingAddrRequest{
			Remark:           remark,
			RecipientAddress: s.eosCfg.Exapp.VaultEVMAddress,
		})
		if err != nil {
			return entity.RespFirstDeposit{}, err
		}
		log.Printf("mapping new btc address txid: %v", resp.TransactionID)

		newDepositAddress, err = s.pollForBTCAddress(ctx, btcBridgeClient, eos.RequestBTCDepositAddress{
			Remark:              remark,
			RecipientEVMAddress: s.eosCfg.Exapp.VaultEVMAddress,
		})
		if err != nil {
			log.Printf("get btc deposit address from bridge error: %v", err)
			return entity.RespFirstDeposit{}, err
		}
	} else {
		bridgeClient := eos.NewBridgeClient(
			s.eosCfg.NodeURL,
			s.eosCfg.Exsat.BridgeContract,
			s.eosCfg.Exapp.Actor,
			s.eosCfg.Exapp.ActorPrivateKey,
		)

		resp, err := bridgeClient.MappingAddress(ctx, eos.MappingAddrRequest{
			PermissionID:     targetChain.PermissionID,
			RecipientAddress: s.eosCfg.Exapp.VaultEVMAddress,
			Remark:           remark,
		})
		if err != nil {
			return entity.RespFirstDeposit{}, err
		}
		log.Printf("mapping new address txid: %v", resp.TransactionID)

		newDepositAddress, err = s.pollForDepositAddress(ctx, bridgeClient, eos.RequestDepositAddress{
			PermissionID: targetChain.PermissionID,
			Remark:       remark,
			Recipient:    s.eosCfg.Exapp.VaultEVMAddress,
		})
		if err != nil {
			log.Printf("get deposit address from bridge error: %v", err)
			return entity.RespFirstDeposit{}, err
		}
	}

	log.Printf("get first deposit address: %s", newDepositAddress)
	err = s.repo.CreateUserDepositAddress(ctx, &db.UserDepositAddress{
		UID:          uid,
		Address:      newDepositAddress,
		PermissionID: targetChain.PermissionID,
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

	token, err := s.repo.GetToken(ctx, req.Symbol)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return entity.RespDeposit{}, errors.New("token not found")
		}
		return entity.RespDeposit{}, err
	}
	var targetChain db.ChainInfo
	for _, chain := range token.Chains {
		if chain.ChainName == req.ChainName {
			targetChain = chain
			break
		}
	}
	if targetChain.ChainName == "" {
		return entity.RespDeposit{}, errors.New("chain not found")
	}

	depositAddress, err := s.repo.GetUserDepositAddress(ctx, uid, targetChain.PermissionID)
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

	var newDepositAddress string
	if req.Symbol == "BTC" && req.ChainName != "eos" && req.ChainName != "exsat" {
		btcBridgeClient := eos.NewBTCBridgeClient(
			s.eosCfg.NodeURL,
			s.eosCfg.Exsat.BTCBridgeContract,
			s.eosCfg.Exapp.Actor,
			s.eosCfg.Exapp.ActorPrivateKey,
		)

		resp, err := btcBridgeClient.MappingAddress(ctx, eos.BTCMappingAddrRequest{
			Remark:           remark,
			RecipientAddress: s.eosCfg.Exapp.VaultEVMAddress,
		})
		if err != nil {
			return entity.RespDeposit{}, err
		}
		log.Printf("mapping new btc address txid: %v", resp.TransactionID)

		newDepositAddress, err = s.pollForBTCAddress(ctx, btcBridgeClient, eos.RequestBTCDepositAddress{
			Remark:              remark,
			RecipientEVMAddress: s.eosCfg.Exapp.VaultEVMAddress,
		})
		if err != nil {
			log.Printf("get btc deposit address from bridge error: %v", err)
			return entity.RespDeposit{}, err
		}
	} else {
		bridgeClient := eos.NewBridgeClient(
			s.eosCfg.NodeURL,
			s.eosCfg.Exsat.BridgeContract,
			s.eosCfg.Exapp.Actor,
			s.eosCfg.Exapp.ActorPrivateKey,
		)

		resp, err := bridgeClient.MappingAddress(ctx, eos.MappingAddrRequest{
			PermissionID:     targetChain.PermissionID,
			RecipientAddress: s.eosCfg.Exapp.VaultEVMAddress,
			Remark:           remark,
		})
		if err != nil {
			return entity.RespDeposit{}, err
		}
		log.Printf("mapping new address txid: %v", resp.TransactionID)

		newDepositAddress, err = s.pollForDepositAddress(ctx, bridgeClient, eos.RequestDepositAddress{
			PermissionID: targetChain.PermissionID,
			Remark:       remark,
			Recipient:    s.eosCfg.Exapp.VaultEVMAddress,
		})
		if err != nil {
			log.Printf("get deposit address from bridge error: %v", err)
			return entity.RespDeposit{}, err
		}
	}

	log.Printf("new deposit address: %s", newDepositAddress)
	err = s.repo.CreateUserDepositAddress(ctx, &db.UserDepositAddress{
		UID:          uid,
		Address:      newDepositAddress,
		PermissionID: targetChain.PermissionID,
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
