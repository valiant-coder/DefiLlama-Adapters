package syncer

import (
	"context"
	"exapp-go/pkg/hyperion"
	"fmt"
	"log"
	"strconv"
	"time"
)

func (s *Service) initPoolLastBlockNum(ctx context.Context) error {
	lastBlockNum, err := s.repo.GetPoolMaxUpdateBlockNum(ctx)
	if err != nil {
		return fmt.Errorf("get max update block number failed: %w", err)
	}

	if lastBlockNum == 0 {
		lastBlockNum = s.hyperionCfg.StartBlock
	} else {
		lastBlockNum = lastBlockNum + 1
	}
	s.poolLastBlockNum = lastBlockNum
	return nil
}

func (s *Service) syncPoolHistory(ctx context.Context) error {
	if err := s.initPoolLastBlockNum(ctx); err != nil {
		return fmt.Errorf("init pool last block number failed: %w", err)
	}

	for {
		resp, err := s.hyperionClient.GetActions(ctx, hyperion.GetActionsRequest{
			Account: "",
			Filter:  fmt.Sprintf(
				"%s:%s,%s:%s", 
				s.cdexCfg.PoolContract, s.eosCfg.Events.Create, 
				s.cdexCfg.PoolContract, s.eosCfg.Events.SetMinAmt,
			),
			Limit:   s.hyperionCfg.BatchSize,
			Sort:    "asc",
			After:   strconv.FormatUint(s.poolLastBlockNum, 10),
		})
		if err != nil {
			return fmt.Errorf("get actions failed: %w", err)
		}

		if len(resp.Actions) == 0 {
			break
		}
		log.Printf("sync pool actions history: %d", len(resp.Actions))

		for _, action := range resp.Actions {
			if err := s.publishAction(action); err != nil {
				return fmt.Errorf("publish action failed: %w", err)
			}
			s.poolLastBlockNum = action.BlockNum
		}
		if len(resp.Actions) < s.hyperionCfg.BatchSize {
			break
		}

		time.Sleep(time.Millisecond * 100)
	}
	return nil
}


func (s *Service) SyncPool(ctx context.Context) (<-chan hyperion.Action, error) {
	if err := s.syncPoolHistory(ctx); err != nil {
		return nil, fmt.Errorf("sync pool history failed: %w", err)
	}
	log.Printf("sync pool history done, last block number: %d", s.poolLastBlockNum)
	poolActionCh, err := s.streamClient.SubscribeAction([]hyperion.ActionStreamRequest{
		{
			Contract:  s.cdexCfg.PoolContract,
			Action:    s.eosCfg.Events.Create,
			Account:   "",
			StartFrom: int64(s.poolLastBlockNum) + 1,
			ReadUntil: 0,
			Filters:   []hyperion.RequestFilter{},
		},
		{
			Contract:  s.cdexCfg.PoolContract,
			Action:    s.eosCfg.Events.SetMinAmt,
			Account:   "",
			StartFrom: int64(s.poolLastBlockNum) + 1,
			ReadUntil: 0,
			Filters:   []hyperion.RequestFilter{},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("subscribe pool action failed: %w", err)
	}
	return poolActionCh, nil
}
