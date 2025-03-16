package syncer

import (
	"context"
	"exapp-go/pkg/hyperion"
	"fmt"
	"log"
	"strconv"
)

func (s *Service) initTokenChainLastBlockNum(ctx context.Context) error {
	lastTokenBlockNum, err := s.repo.GetTokenMaxBlockNum(ctx)
	if err != nil {
		return err
	}
	chainLastBlockNum, err := s.repo.GetChainMaxBlockNum(ctx)
	if err != nil {
		return err
	}

	if lastTokenBlockNum == 0 && chainLastBlockNum == 0 {
		lastTokenBlockNum = s.hyperionCfg.StartBlock
	} else {
		if lastTokenBlockNum > chainLastBlockNum {
			lastTokenBlockNum = lastTokenBlockNum + 1
		} else {
			lastTokenBlockNum = chainLastBlockNum + 1
		}
	}
	s.tokenChainLastBlockNum = lastTokenBlockNum
	return nil
}

func (s *Service) syncTokenChainHistories(ctx context.Context) error {
	if err := s.initTokenChainLastBlockNum(ctx); err != nil {
		return fmt.Errorf("init token last block number failed: %w", err)
	}

	for {
		resp, err := s.hyperionClient.GetActions(ctx, hyperion.GetActionsRequest{
			Account: "",
			Filter: fmt.Sprintf(
				"%s:%s,%s:%s",
				s.oneDexCfg.BridgeContract, s.eosCfg.Events.CreateToken,
				s.exsatCfg.BridgeContract, s.eosCfg.Events.AddXSATChain),
			Limit: s.hyperionCfg.BatchSize,
			Sort:  "asc",
			After: strconv.FormatUint(s.tokenChainLastBlockNum, 10),
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
			s.tokenChainLastBlockNum = action.BlockNum
		}
		if len(resp.Actions) < s.hyperionCfg.BatchSize {
			break
		}
	}
	return nil
}

func (s *Service) SyncTokenChain(ctx context.Context) (<-chan hyperion.Action, error) {
	if err := s.syncTokenChainHistories(ctx); err != nil {
		return nil, fmt.Errorf("sync token histories failed: %w", err)
	}
	log.Printf("sync token histories done, last block number: %d", s.tokenChainLastBlockNum)
	tokenActionCh, err := s.streamClient.SubscribeAction([]hyperion.ActionStreamRequest{
		{
			Contract:  s.oneDexCfg.BridgeContract,
			Action:    s.eosCfg.Events.CreateToken,
			Account:   "",
			StartFrom: int64(s.tokenChainLastBlockNum) + 1,
			ReadUntil: 0,
			Filters:   []hyperion.RequestFilter{},
		},
		{
			Contract:  s.exsatCfg.BridgeContract,
			Action:    s.eosCfg.Events.AddXSATChain,
			Account:   "",
			StartFrom: int64(s.tokenChainLastBlockNum) + 1,
			ReadUntil: 0,
			Filters:   []hyperion.RequestFilter{},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("subscribe token action failed: %w", err)
	}
	return tokenActionCh, nil
}
