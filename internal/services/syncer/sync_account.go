package syncer

import (
	"context"
	"exapp-go/pkg/hyperion"
	"fmt"
	"log"
	"strconv"
	"time"
)

func (s *Service) initAccountLastBlockNum(ctx context.Context) error {
	lastBlockNum, err := s.repo.GetUserCredentialMaxBlockNumber(ctx)
	if err != nil {
		return fmt.Errorf("get user credential max block number failed: %w", err)
	}

	if lastBlockNum == 0 {
		lastBlockNum = s.hyperionCfg.StartBlock
	} else {
		lastBlockNum = lastBlockNum + 1
	}
	s.accountLastBlockNum = lastBlockNum

	return nil
}

func (s *Service) syncAccountHistory(ctx context.Context) error {
	if err := s.initAccountLastBlockNum(ctx); err != nil {
		return fmt.Errorf("init account last block number failed: %w", err)
	}

	for {
		resp, err := s.hyperionClient.GetActions(ctx, hyperion.GetActionsRequest{
			Account: "",
			Filter:  "eosio:updateauth",
			Limit:   s.hyperionCfg.BatchSize,
			Sort:    "asc",
			After:   strconv.FormatUint(s.accountLastBlockNum, 10),
		})
		if err != nil {
			return fmt.Errorf("get actions failed: %w", err)
		}

		if len(resp.Actions) == 0 {
			break
		}

		for _, action := range resp.Actions {
			if err := s.publishAction(action); err != nil {
				return fmt.Errorf("publish action failed: %w", err)
			}

			s.accountLastBlockNum = action.BlockNum
		}

		if len(resp.Actions) < s.hyperionCfg.BatchSize {
			break
		}

		time.Sleep(time.Millisecond * 100)
	}

	return nil
}

func (s *Service) SyncAccount(ctx context.Context) (<-chan hyperion.Action, error) {
	if err := s.syncAccountHistory(ctx); err != nil {
		return nil, fmt.Errorf("sync account history failed: %w", err)
	}
	log.Printf("sync account history done, last block number: %d", s.accountLastBlockNum)
	accountActionCh, err := s.streamClient.SubscribeAction([]hyperion.ActionStreamRequest{
		{
			Contract:  "eosio",
			Action:    "updateauth",
			Account:   "",
			StartFrom: int64(s.accountLastBlockNum) + 1,
			ReadUntil: 0,
			Filters:   []hyperion.RequestFilter{},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("subscribe account action failed: %w", err)
	}

	return accountActionCh, nil
}
