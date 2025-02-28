package syncer

import (
	"context"
	"exapp-go/internal/db/db"
	"exapp-go/pkg/hyperion"
	"fmt"
	"log"
	"strconv"
)

func (s *Service) initTradeLastBlockNum(ctx context.Context) error {

	lastBlockNum, err := s.ckhRepo.GetTradeMaxBlockNumber(ctx)
	if err != nil {
		return fmt.Errorf("get trade max block number failed: %w", err)
	}

	repo := db.New()
	lastOpenOrderBlockNum, err := repo.GetOpenOrderMaxBlockNumber(context.Background())
	if err != nil {
		return fmt.Errorf("get open order max block number failed: %w", err)
	}

	if lastOpenOrderBlockNum > lastBlockNum {
		lastBlockNum = lastOpenOrderBlockNum
	}

	if lastBlockNum == 0 {
		lastBlockNum = s.hyperionCfg.StartBlock
	} else {
		lastBlockNum = lastBlockNum + 1
	}
	s.tradeLastBlockNum = int64(lastBlockNum)

	return nil
}

func (s *Service) syncTradeHistories(ctx context.Context) error {
	if err := s.initTradeLastBlockNum(ctx); err != nil {
		return fmt.Errorf("init trade last block number failed: %w", err)
	}

	for {
		resp, err := s.hyperionClient.GetActions(ctx, hyperion.GetActionsRequest{
			Account: "",
			Filter: fmt.Sprintf(
				"%s:%s,%s:%s,%s:%s",
				s.cdexCfg.EventContract, s.eosCfg.Events.EmitPlaced,
				s.cdexCfg.EventContract, s.eosCfg.Events.EmitCanceled,
				s.cdexCfg.EventContract, s.eosCfg.Events.EmitFilled,
			),
			Limit: s.hyperionCfg.BatchSize,
			Sort:  "asc",
			After: strconv.FormatInt(s.tradeLastBlockNum, 10),
		})
		if err != nil {
			return fmt.Errorf("get actions failed: %w", err)
		}

		if len(resp.Actions) == 0 {
			break
		}
		log.Printf("sync trade actions history: %d", len(resp.Actions))

		for _, action := range resp.Actions {
			if err := s.publishAction(action); err != nil {
				return fmt.Errorf("publish action failed: %w", err)
			}

			s.tradeLastBlockNum = int64(action.BlockNum)
		}

		if len(resp.Actions) < s.hyperionCfg.BatchSize {
			break
		}

		// time.Sleep(time.Millisecond * 100)
	}

	return nil
}

func (s *Service) SyncTrade(ctx context.Context) (<-chan hyperion.Action, error) {
	if s.syncTradeHistory {
		if err := s.syncTradeHistories(ctx); err != nil {
			return nil, fmt.Errorf("sync trade history failed: %w", err)
		}
		log.Printf("sync trade history done, last block number: %d", s.tradeLastBlockNum)
	}

	actionCh, err := s.streamClient.SubscribeAction([]hyperion.ActionStreamRequest{

		{
			Contract:  s.cdexCfg.EventContract,
			Action:    s.eosCfg.Events.EmitPlaced,
			Account:   "",
			StartFrom: s.tradeLastBlockNum+ 1,
			ReadUntil: 0,
			Filters:   []hyperion.RequestFilter{},
		},
		{
			Contract:  s.cdexCfg.EventContract,
			Action:    s.eosCfg.Events.EmitCanceled,
			Account:   "",
			StartFrom: s.tradeLastBlockNum + 1,
			ReadUntil: 0,
			Filters:   []hyperion.RequestFilter{},
		},
		{
			Contract:  s.cdexCfg.EventContract,
			Action:    s.eosCfg.Events.EmitFilled,
			Account:   "",
			StartFrom: s.tradeLastBlockNum + 1,
			ReadUntil: 0,
			Filters:   []hyperion.RequestFilter{},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("subscribe trade action failed: %w", err)
	}
	return actionCh, nil
}
