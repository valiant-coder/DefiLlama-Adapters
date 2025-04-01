package syncer

import (
	"context"
	"exapp-go/pkg/hyperion"
	"fmt"
	"log"
	"strconv"
)

func (s *Service) initDepositLastBlockNum(ctx context.Context) error {
	lastBlockNum, err := s.repo.GetDepositMaxBlockNumber(ctx)
	if err != nil {
		return fmt.Errorf("get deposit max block number failed: %w", err)
	}

	if lastBlockNum == 0 {
		lastBlockNum = s.hyperionCfg.StartBlock
	} else {
		lastBlockNum = lastBlockNum + 1
	}
	s.depositLastBlockNum = lastBlockNum
	return nil
}

func (s *Service) syncDepositHistory(ctx context.Context) error {
	if err := s.initDepositLastBlockNum(ctx); err != nil {
		return fmt.Errorf("init deposit last block number failed: %w", err)
	}

	for {
		resp, err := s.hyperionClient.GetActions(ctx, hyperion.GetActionsRequest{
			Account: "",
			Filter: fmt.Sprintf(
				"%s:%s,%s:%s,%s:%s,%s:%s",
				s.oneDexCfg.SignUpContract, s.eosCfg.Events.LogNewAcc,
				s.oneDexCfg.PortalContract, s.eosCfg.Events.LogDeposit,
				s.exsatCfg.BridgeContract, s.eosCfg.Events.DepositLog,
				s.exsatCfg.BTCBridgeContract, s.eosCfg.Events.DepositLog,
			),
			Limit: s.hyperionCfg.BatchSize,
			Sort:  "asc",
			After: strconv.FormatUint(s.depositLastBlockNum, 10),
		})
		if err != nil {
			return fmt.Errorf("get actions failed: %w", err)
		}

		if len(resp.Actions) == 0 {
			break
		}
		log.Printf("sync deposit actions history: %d", len(resp.Actions))

		for _, action := range resp.Actions {
			if err := s.publishAction(action); err != nil {
				return fmt.Errorf("publish action failed: %w", err)
			}

			s.depositLastBlockNum = action.BlockNum
		}

		if len(resp.Actions) < s.hyperionCfg.BatchSize {
			break
		}

		// time.Sleep(time.Millisecond * 100)
	}

	return nil
}

func (s *Service) SyncDeposit(ctx context.Context) (<-chan hyperion.Action, error) {
	if err := s.syncDepositHistory(ctx); err != nil {
		return nil, fmt.Errorf("sync deposit history failed: %w", err)
	}
	log.Printf("sync deposit history done, last block number: %d", s.depositLastBlockNum)
	depositActionCh, err := s.streamClient.SubscribeAction([]hyperion.ActionStreamRequest{
		{
			Contract:  s.exsatCfg.BridgeContract,
			Action:    s.eosCfg.Events.DepositLog,
			Account:   "",
			StartFrom: int64(s.depositLastBlockNum) + 1,
			ReadUntil: 0,
			Filters:   []hyperion.RequestFilter{},
		},
		{
			Contract:  s.oneDexCfg.SignUpContract,
			Action:    s.eosCfg.Events.LogNewAcc,
			Account:   "",
			StartFrom: int64(s.depositLastBlockNum) + 1,
			ReadUntil: 0,
			Filters:   []hyperion.RequestFilter{},
		},
		{
			Contract:  s.oneDexCfg.PortalContract,
			Action:    s.eosCfg.Events.LogDeposit,
			Account:   "",
			StartFrom: int64(s.depositLastBlockNum) + 1,
			ReadUntil: 0,
			Filters:   []hyperion.RequestFilter{},
		},

		{
			Contract:  s.exsatCfg.BTCBridgeContract,
			Action:    s.eosCfg.Events.DepositLog,
			Account:   "",
			StartFrom: int64(s.depositLastBlockNum) + 1,
			ReadUntil: 0,
			Filters:   []hyperion.RequestFilter{},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("subscribe deposit action failed: %w", err)
	}

	return depositActionCh, nil
}
