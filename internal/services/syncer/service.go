package syncer

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"exapp-go/config"
	"exapp-go/internal/db/ckhdb"
	"exapp-go/pkg/hyperion"
	"exapp-go/pkg/nsqutil"
)

const (
	TopicActionSync = "cdex_action_sync"
)

type Service struct {
	hyperionClient *hyperion.Client
	streamClient   *hyperion.StreamClient
	publisher      *nsqutil.Client
	lastBlockNum   uint64
	hyperionCfg    config.HyperionConfig
	nsqCfg         config.NsqConfig
	cdexCfg        config.CdexConfig
}

func NewService(hyperionCfg config.HyperionConfig, nsqCfg config.NsqConfig, cdexCfg config.CdexConfig) (*Service, error) {
	hyperionClient := hyperion.NewClient(hyperionCfg.Endpoint)
	streamClient, err := hyperion.NewStreamClient(hyperionCfg.StreamEndpoint)
	if err != nil {
		return nil, fmt.Errorf("create stream client failed: %w", err)
	}

	ckhRepo := ckhdb.New()
	lastBlockNum, err := ckhRepo.GetMaxBlockNumber(context.Background())
	if err != nil {
		return nil, fmt.Errorf("get max block number failed: %w", err)
	}
	if lastBlockNum == 0 {
		lastBlockNum = hyperionCfg.StartBlock
	}
	return &Service{
		hyperionClient: hyperionClient,
		streamClient:   streamClient,
		publisher:      nsqutil.NewPublisher(nsqCfg.Nsqds),
		lastBlockNum:   lastBlockNum,
		hyperionCfg:    hyperionCfg,
		nsqCfg:         nsqCfg,
		cdexCfg:        cdexCfg,
	}, nil
}

func (s *Service) Start(ctx context.Context) error {
	if err := s.syncHistory(ctx); err != nil {
		return fmt.Errorf("sync history failed: %w", err)
	}

	actionCh, err := s.streamClient.SubscribeAction([]hyperion.ActionStreamRequest{
		{
			Contract: s.cdexCfg.PoolContract,
			StartFrom: int64(s.hyperionCfg.StartBlock),

		},
		{
			Contract: s.cdexCfg.EventContract,
			StartFrom: int64(s.hyperionCfg.StartBlock),
		},
	})

	if err != nil {
		return fmt.Errorf("subscribe actions failed: %w", err)
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case action, ok := <-actionCh:
			if !ok {
				return fmt.Errorf("action channel closed")
			}

			if err := s.publishAction(action); err != nil {
				log.Printf("Publish action failed: %v", err)
				continue
			}

			s.lastBlockNum = action.BlockNum
		}
	}
}

func (s *Service) Stop() error {

	s.publisher.Stop()
	return s.streamClient.Close()
}

func (s *Service) syncHistory(ctx context.Context) error {
	for {
		resp, err := s.hyperionClient.GetActions(ctx, hyperion.GetActionsRequest{
			Account: "",
			Filter:  fmt.Sprintf("%s:*,%s:*",s.cdexCfg.PoolContract, s.cdexCfg.EventContract),
			Limit:   s.hyperionCfg.BatchSize,
			Sort:    "asc",
			After:   strconv.FormatUint(s.lastBlockNum, 10),
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

			s.lastBlockNum = action.BlockNum
		}

		if len(resp.Actions) < s.hyperionCfg.BatchSize {
			break
		}

		time.Sleep(time.Millisecond * 100)
	}

	return nil
}

func (s *Service) publishAction(action hyperion.Action) error {
	return s.publisher.Publish(TopicActionSync, action)
}
