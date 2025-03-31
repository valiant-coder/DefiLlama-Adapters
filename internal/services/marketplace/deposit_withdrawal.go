package marketplace

import (
	"context"
	"errors"
	"exapp-go/config"
	"exapp-go/internal/db/db"
	"exapp-go/internal/entity"
	"exapp-go/internal/errno"
	"exapp-go/pkg/eos/onedex"
	"exapp-go/pkg/queryparams"
	"fmt"
	"log"
	"time"

	"github.com/spf13/cast"
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

func (s *DepositWithdrawalService) pollForDepositAddress(ctx context.Context, bridgeClient *onedex.BridgeClient, req onedex.RequestDepositAddress) (string, error) {
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

func (s *DepositWithdrawalService) pollForBTCAddress(ctx context.Context, bridgeClient *onedex.BTCBridgeClient, req onedex.RequestBTCDepositAddress) (string, error) {
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

func (s *DepositWithdrawalService) Deposit(ctx context.Context, uid string, req *entity.ReqDeposit) (entity.RespDeposit, error) {
	passkey, err := s.repo.GetUserCredential(ctx, uid)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return entity.RespDeposit{}, errno.DefaultParamsError("not found passkey")
		}
		return entity.RespDeposit{}, err
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
		if chain.ChainID == req.ChainID {
			targetChain = chain
			break
		}
	}

	remark := fmt.Sprintf("topup-%s", uid)
	depositAddress, err := s.repo.GetUserDepositAddress(ctx, uid, targetChain.PermissionID)
	if err != nil {
		return entity.RespDeposit{}, err
	}
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
	if req.Symbol == "BTC" && targetChain.DepositByBTCBridge {
		btcBridgeClient := onedex.NewBTCBridgeClient(
			s.eosCfg.NodeURL,
			s.eosCfg.Exsat.BTCBridgeContract,
			s.eosCfg.OneDex.Actor,
			s.eosCfg.OneDex.ActorPrivateKey,
			s.eosCfg.OneDex.ActorPermission,
		)

		resp, err := btcBridgeClient.MappingAddress(ctx, onedex.BTCMappingAddrRequest{
			Remark:           remark,
			RecipientAddress: s.eosCfg.Exsat.BridgeExtensionEVMAddress,
		})
		if err != nil {
			return entity.RespDeposit{}, err
		}
		log.Printf("mapping new btc address txid: %v", resp.TransactionID)

		newDepositAddress, err = s.pollForBTCAddress(ctx, btcBridgeClient, onedex.RequestBTCDepositAddress{
			Remark:              remark,
			RecipientEVMAddress: s.eosCfg.Exsat.BridgeExtensionEVMAddress,
		})
		if err != nil {
			log.Printf("get btc deposit address from bridge error: %v", err)
			return entity.RespDeposit{}, err
		}

	} else {
		bridgeClient := onedex.NewBridgeClient(
			s.eosCfg.NodeURL,
			s.eosCfg.Exsat.BridgeContract,
			s.eosCfg.OneDex.Actor,
			s.eosCfg.OneDex.ActorPrivateKey,
			s.eosCfg.OneDex.ActorPermission,
		)

		resp, err := bridgeClient.MappingAddress(ctx, onedex.MappingAddrRequest{
			PermissionID:     targetChain.PermissionID,
			RecipientAddress: s.eosCfg.Exsat.BridgeExtensionEVMAddress,
			Remark:           remark,
		})
		if err != nil {
			return entity.RespDeposit{}, err
		}
		log.Printf("mapping new address txid: %v", resp.TransactionID)

		newDepositAddress, err = s.pollForDepositAddress(ctx, bridgeClient, onedex.RequestDepositAddress{
			PermissionID: targetChain.PermissionID,
			Remark:       remark,
			Recipient:    s.eosCfg.Exsat.BridgeExtensionEVMAddress,
		})
		if err != nil {
			log.Printf("get deposit address from bridge error: %v", err)
			return entity.RespDeposit{}, err
		}
	}
	log.Printf("new deposit address: %s", newDepositAddress)
	signupClient := onedex.NewSignupClient(
		s.eosCfg.NodeURL,
		s.eosCfg.OneDex.Actor,
		s.eosCfg.OneDex.ActorPrivateKey,
		s.eosCfg.OneDex.ActorPermission,
	)
	go func() {
		resp, err := signupClient.ApplyAcc(ctx, cast.ToUint64(uid), passkey.PublicKey)
		if err != nil {
			log.Printf("apply acc txid: %v", resp.TransactionID)
		}
		log.Printf("apply acc txid: %v", resp.TransactionID)
	}()

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
