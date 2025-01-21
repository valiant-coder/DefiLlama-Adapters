package syncer

import (
	"context"
	"exapp-go/internal/db/db"
	"exapp-go/pkg/hyperion"
	"fmt"
	"log"
	"strconv"
	"time"
)

func (s *Service) initTradeLastBlockNum(ctx context.Context) error {

	lastBlockNum, err := s.ckhRepo.GetMaxBlockNumber(ctx)
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
	s.tradeLastBlockNum = lastBlockNum

	return nil
}

func (s *Service) syncTradeHistory(ctx context.Context) error {
	if err := s.initTradeLastBlockNum(ctx); err != nil {
		return fmt.Errorf("init trade last block number failed: %w", err)
	}

	for {
		resp, err := s.hyperionClient.GetActions(ctx, hyperion.GetActionsRequest{
			Account: "",
			Filter: fmt.Sprintf(
				"%s:create,%s:emitplaced,%s:emitcanceled,%s:emitfilled",
				s.cdexCfg.PoolContract,
				s.cdexCfg.EventContract,
				s.cdexCfg.EventContract,
				s.cdexCfg.EventContract,
			),
			Limit: s.hyperionCfg.BatchSize,
			Sort:  "asc",
			After: strconv.FormatUint(s.tradeLastBlockNum, 10),
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

			s.tradeLastBlockNum = action.BlockNum
		}

		if len(resp.Actions) < s.hyperionCfg.BatchSize {
			break
		}

		time.Sleep(time.Millisecond * 100)
	}

	return nil
}

func (s *Service) SyncTrade(ctx context.Context) (<-chan hyperion.Action, error) {

	if err := s.syncTradeHistory(ctx); err != nil {
		return nil, fmt.Errorf("sync trade history failed: %w", err)
	}

	actionCh, err := s.streamClient.SubscribeAction([]hyperion.ActionStreamRequest{
		{
			Contract:  s.cdexCfg.PoolContract,
			Action:    "create",
			Account:   "",
			StartFrom: int64(s.tradeLastBlockNum) + 1,
			ReadUntil: 0,
			Filters:   []hyperion.RequestFilter{},
		},
		{
			Contract:  s.cdexCfg.EventContract,
			Action:    "emitplaced",
			Account:   "",
			StartFrom: int64(s.tradeLastBlockNum) + 1,
			ReadUntil: 0,
			Filters:   []hyperion.RequestFilter{},
		},
		{
			Contract:  s.cdexCfg.EventContract,
			Action:    "emitcanceled",
			Account:   "",
			StartFrom: int64(s.tradeLastBlockNum) + 1,
			ReadUntil: 0,
			Filters:   []hyperion.RequestFilter{},
		},
		{
			Contract:  s.cdexCfg.EventContract,
			Action:    "emitfilled",
			Account:   "",
			StartFrom: int64(s.tradeLastBlockNum) + 1,
			ReadUntil: 0,
			Filters:   []hyperion.RequestFilter{},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("subscribe trade action failed: %w", err)
	}
	return actionCh, nil
}
