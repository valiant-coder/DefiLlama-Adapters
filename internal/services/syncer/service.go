package syncer

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"exapp-go/config"
	"exapp-go/internal/db/ckhdb"
	"exapp-go/internal/db/db"
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
	exappCfg       config.ExappConfig
	exsatCfg       config.ExsatConfig
}

func NewService(eosCfg config.EosConfig, nsqCfg config.NsqConfig) (*Service, error) {
	hyperionCfg := eosCfg.Hyperion
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

	repo := db.New()
	lastOpenOrderBlockNum, err := repo.GetOpenOrderMaxBlockNumber(context.Background())
	if err != nil {
		return nil, fmt.Errorf("get max block number failed: %w", err)
	}

	if lastOpenOrderBlockNum > lastBlockNum {
		lastBlockNum = lastOpenOrderBlockNum
	}

	if lastBlockNum == 0 {
		lastBlockNum = hyperionCfg.StartBlock
	} else {
		lastBlockNum = lastBlockNum + 1
	}

	return &Service{
		hyperionClient: hyperionClient,
		streamClient:   streamClient,
		publisher:      nsqutil.NewPublisher(nsqCfg.Nsqds),
		lastBlockNum:   lastBlockNum,
		hyperionCfg:    hyperionCfg,
		nsqCfg:         nsqCfg,
		cdexCfg:        eosCfg.CdexConfig,
		exappCfg:       eosCfg.Exapp,
	}, nil
}

func (s *Service) Start(ctx context.Context) error {
	if err := s.syncHistory(ctx); err != nil {
		return fmt.Errorf("sync history failed: %w", err)
	}

	actionCh, err := s.streamClient.SubscribeAction([]hyperion.ActionStreamRequest{
		{
			Contract:  s.cdexCfg.PoolContract,
			Action:    "*",
			Account:   "",
			StartFrom: int64(s.lastBlockNum) + 1,
			ReadUntil: 0,
			Filters:   []hyperion.RequestFilter{},
		},
		{
			Contract:  s.cdexCfg.EventContract,
			Action:    "*",
			Account:   "",
			StartFrom: int64(s.lastBlockNum) + 1,
			ReadUntil: 0,
			Filters:   []hyperion.RequestFilter{},
		},
		{
			Contract:  s.exappCfg.AssetContract,
			Action:    "*",
			Account:   "",
			StartFrom: int64(s.lastBlockNum) + 1,
			ReadUntil: 0,
			Filters:   []hyperion.RequestFilter{},
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
			log.Printf("new action: %v", string(action.TrxID))
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
			Filter:  fmt.Sprintf("%s:*,%s:*,%s:*", s.cdexCfg.PoolContract, s.cdexCfg.EventContract, s.exappCfg.AssetContract),
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
		log.Printf("sync actions history: %d", len(resp.Actions))

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
