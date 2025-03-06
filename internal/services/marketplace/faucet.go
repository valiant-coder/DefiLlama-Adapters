package marketplace

import (
	"context"
	"exapp-go/config"
	"exapp-go/internal/db/db"
	"exapp-go/internal/entity"
	"exapp-go/internal/errno"
	"exapp-go/pkg/eth"
	"log"

	"github.com/shopspring/decimal"
)

func NewFaucetService() *FaucetService {
	return &FaucetService{
		repo: db.New(),
	}
}

type FaucetService struct {
	repo *db.Repo
}

func (s *FaucetService) ClaimFaucet(ctx context.Context, req *entity.ReqClaimFaucet) (*entity.RespClaimFaucet, error) {
	faucetConfig := config.Conf().Faucet
	if !faucetConfig.Enabled {
		return nil, errno.NewParamsError(errno.EpFaucetNotEnabled)
	}
	uid, err := s.repo.GetUIDByDepositAddress(ctx, req.DepositAddress)
	if err != nil {
		return nil, err
	}
	if uid == "" {
		return nil, errno.NewParamsError(errno.EpFaucetNotRegistered)
	}
	isClaimed, err := s.repo.IsUserClaimFaucet(ctx, uid)
	if err != nil {
		return nil, err
	}
	if isClaimed {
		return nil, errno.NewParamsError(errno.EpFaucetClaimed)
	}
	record := &db.FaucetRecord{
		UID:            uid,
		DepositAddress: req.DepositAddress,
		Token:          "USDT",
		Amount:         decimal.NewFromFloat(faucetConfig.Amount),
	}
	ethClient, err := eth.NewClient(faucetConfig.EVMRpcUrl)
	if err != nil {
		return nil, err
	}
	tokenClient, err := eth.NewToken(faucetConfig.USDTTokenAddress)
	if err != nil {
		return nil, err
	}

	txHash, err := tokenClient.MintERC20(
		ethClient,
		faucetConfig.TokenOwner,
		req.DepositAddress,
		faucetConfig.Amount,
		faucetConfig.TokenOwnerPrivateKey,
		0,
	)
	if err != nil {
		log.Default().Printf("transfer usdt error: %v", err)
		return nil, errno.DefaultParamsError("Transfer usdt error")
	}
	record.TxHash = txHash
	err = s.repo.CreateFaucetRecord(ctx, record)
	if err != nil {
		return nil, err
	}

	return &entity.RespClaimFaucet{
		TxHash: txHash,
	}, nil
}
