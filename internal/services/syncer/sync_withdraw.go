package syncer

import (
	"context"
	"exapp-go/pkg/hyperion"
	"fmt"
	"log"
	"strconv"
)

func (s *Service) initWithdrawLastBlockNum(ctx context.Context) error {
	lastBlockNum, err := s.repo.GetWithdrawMaxBlockNumber(ctx)
	if err != nil {
		return fmt.Errorf("get withdraw max block number failed: %w", err)
	}

	if lastBlockNum == 0 {
		lastBlockNum = s.hyperionCfg.StartBlock
	} else {
		lastBlockNum = lastBlockNum + 1
	}
	s.withdrawLastBlockNum = lastBlockNum
	return nil
}

func (s *Service) syncWithdrawHistory(ctx context.Context) error {
	if err := s.initWithdrawLastBlockNum(ctx); err != nil {
		return fmt.Errorf("init withdraw last block number failed: %w", err)
	}

	for {
		resp, err := s.hyperionClient.GetActions(ctx, hyperion.GetActionsRequest{
			Account: "",
			Filter: fmt.Sprintf(
				"%s:%s,%s:%s,%s:%s,%s:%s",
				s.exappCfg.BridgeContract, s.eosCfg.Events.LogWithdraw,
				s.exsatCfg.BridgeContract, s.eosCfg.Events.WithdrawLog,
				s.exsatCfg.BTCBridgeContract, s.eosCfg.Events.WithdrawLog,
				s.exappCfg.BridgeContract, s.eosCfg.Events.LogSend,
			),
			Limit: s.hyperionCfg.BatchSize,
			Sort:  "asc",
			After: strconv.FormatUint(s.withdrawLastBlockNum, 10),
		})
		if err != nil {
			return fmt.Errorf("get actions failed: %w", err)
		}

		if len(resp.Actions) == 0 {
			break
		}
		log.Printf("sync withdraw actions history: %d", len(resp.Actions))

		for _, action := range resp.Actions {
			if err := s.publishAction(action); err != nil {
				return fmt.Errorf("publish action failed: %w", err)
			}

			s.withdrawLastBlockNum = action.BlockNum
		}

		if len(resp.Actions) < s.hyperionCfg.BatchSize {
			break
		}
	}

	return nil
}

func (s *Service) SyncWithdraw(ctx context.Context) (<-chan hyperion.Action, error) {
	if err := s.syncWithdrawHistory(ctx); err != nil {
		return nil, fmt.Errorf("sync withdraw history failed: %w", err)
	}
	log.Printf("sync withdraw history done, last block number: %d", s.withdrawLastBlockNum)
	actionCh, err := s.streamClient.SubscribeAction([]hyperion.ActionStreamRequest{
		{
			Contract:  s.exappCfg.BridgeContract,
			Action:    s.eosCfg.Events.LogWithdraw,
			Account:   "",
			StartFrom: int64(s.withdrawLastBlockNum) + 1,
			ReadUntil: 0,
			Filters:   []hyperion.RequestFilter{},
		},
		{
			Contract:  s.exsatCfg.BridgeContract,
			Action:    s.eosCfg.Events.WithdrawLog,
			Account:   "",
			StartFrom: int64(s.withdrawLastBlockNum) + 1,
			ReadUntil: 0,
			Filters:   []hyperion.RequestFilter{},
		},
		{
			Contract:  s.exsatCfg.BTCBridgeContract,
			Action:    s.eosCfg.Events.WithdrawLog,
			Account:   "",
			StartFrom: int64(s.withdrawLastBlockNum) + 1,
			ReadUntil: 0,
			Filters:   []hyperion.RequestFilter{},
		},
		{
			Contract:  s.exappCfg.BridgeContract,
			Action:    s.eosCfg.Events.LogSend,
			Account:   "",
			StartFrom: int64(s.withdrawLastBlockNum) + 1,
			ReadUntil: 0,
			Filters:   []hyperion.RequestFilter{},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("subscribe withdraw action failed: %w", err)
	}
	return actionCh, nil
}
