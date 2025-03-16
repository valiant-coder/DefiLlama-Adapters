package syncer

import (
	"context"
	"exapp-go/pkg/hyperion"
	"fmt"
	"log"
	"strconv"
)

func (s *Service) initTokenLastBlockNum(ctx context.Context) error {
	lastBlockNum, err := s.repo.GetTokenMaxBlockNum(ctx)
	if err != nil {
		return err
	}

	if lastBlockNum == 0 {
		lastBlockNum = s.hyperionCfg.StartBlock
	} else {
		lastBlockNum = lastBlockNum + 1
	}
	s.tokenLastBlockNum = lastBlockNum
	return nil
}

func (s *Service) syncTokenHistories(ctx context.Context) error {
	if err := s.initTokenLastBlockNum(ctx); err != nil {
		return fmt.Errorf("init token last block number failed: %w", err)
	}

	for {
		resp, err := s.hyperionClient.GetActions(ctx, hyperion.GetActionsRequest{
			Account: "",
			Filter:  fmt.Sprintf("%s:%s", s.oneDexCfg.BridgeContract, s.eosCfg.Events.CreateToken),
			Limit:   s.hyperionCfg.BatchSize,
			Sort:    "asc",
			After:   strconv.FormatUint(s.tokenLastBlockNum, 10),
		})
		if err != nil {
			return fmt.Errorf("get actions failed: %w", err)
		}

		if len(resp.Actions) == 0 {
			break
		}
		log.Printf("sync token histories: %d", len(resp.Actions))
		for _, action := range resp.Actions {
			if err := s.publishAction(action); err != nil {
				return fmt.Errorf("publish action failed: %w", err)
			}
			s.tokenLastBlockNum = action.BlockNum
		}
		if len(resp.Actions) < s.hyperionCfg.BatchSize {
			break
		}
	}
	return nil
}


func (s *Service) SyncToken(ctx context.Context) (<-chan hyperion.Action, error) {
	if err := s.syncTokenHistories(ctx); err != nil {
		return nil, fmt.Errorf("sync token histories failed: %w", err)
	}
	log.Printf("sync token histories done, last block number: %d", s.tokenLastBlockNum)
	tokenActionCh, err := s.streamClient.SubscribeAction([]hyperion.ActionStreamRequest{
		{
			Contract:  s.oneDexCfg.BridgeContract,
			Action:    s.eosCfg.Events.CreateToken,
			Account:   "",
			StartFrom: int64(s.tokenLastBlockNum) + 1,
			ReadUntil: 0,
			Filters:   []hyperion.RequestFilter{},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("subscribe token action failed: %w", err)
	}
	return tokenActionCh, nil
}
