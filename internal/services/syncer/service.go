package syncer

import (
	"context"
	"fmt"
	"log"

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
	repo                 *db.Repo
	ckhRepo              *ckhdb.ClickHouseRepo
	hyperionClient       *hyperion.Client
	streamClient         *hyperion.StreamClient
	publisher            *nsqutil.Publisher
	tradeLastBlockNum    int64
	depositLastBlockNum  uint64
	withdrawLastBlockNum uint64
	accountLastBlockNum  uint64
	poolLastBlockNum     uint64
	hyperionCfg          config.HyperionConfig
	nsqCfg               config.NsqConfig
	eosCfg               config.EosConfig
	cdexCfg              config.CdexConfig
	oneDexCfg            config.OneDexConfig
	exsatCfg             config.ExsatConfig
	syncTradeHistory     bool
}

func NewService(eosCfg config.EosConfig, nsqCfg config.NsqConfig) (*Service, error) {
	hyperionCfg := eosCfg.Hyperion
	hyperionClient := hyperion.NewClient(hyperionCfg.Endpoint)
	streamClient, err := hyperion.NewStreamClient(hyperionCfg.StreamEndpoint)
	if err != nil {
		return nil, fmt.Errorf("create stream client failed: %w", err)
	}

	return &Service{
		repo:             db.New(),
		ckhRepo:          ckhdb.New(),
		hyperionClient:   hyperionClient,
		streamClient:     streamClient,
		publisher:        nsqutil.NewPublisher(nsqCfg.Nsqds),
		hyperionCfg:      hyperionCfg,
		nsqCfg:           nsqCfg,
		eosCfg:           eosCfg,
		cdexCfg:          eosCfg.CdexConfig,
		oneDexCfg:        eosCfg.OneDex,
		exsatCfg:         eosCfg.Exsat,
		syncTradeHistory: eosCfg.Hyperion.SyncTradeHistory,
	}, nil
}

func (s *Service) Start(ctx context.Context) error {
	tradeActionsCh, err := s.SyncTrade(ctx)
	if err != nil {
		return fmt.Errorf("sync trade failed: %w", err)
	}
	depositActionsCh, err := s.SyncDeposit(ctx)
	if err != nil {
		return fmt.Errorf("sync deposit failed: %w", err)
	}
	withdrawActionsCh, err := s.SyncWithdraw(ctx)
	if err != nil {
		return fmt.Errorf("sync withdraw failed: %w", err)
	}
	accountActionsCh, err := s.SyncAccount(ctx)
	if err != nil {
		return fmt.Errorf("sync account failed: %w", err)
	}
	poolActionsCh, err := s.SyncPool(ctx)
	if err != nil {
		return fmt.Errorf("sync pool failed: %w", err)
	}
	for {
		select {
		case <-ctx.Done():
			return nil
		case action, ok := <-tradeActionsCh:
			if !ok {
				return fmt.Errorf("trade action channel closed")
			}
			log.Printf("new trade action: %v", string(action.TrxID))
			if err := s.publishAction(action); err != nil {
				log.Printf("Publish trade action failed: %v", err)
				continue
			}
			s.tradeLastBlockNum = int64(action.BlockNum)
		case action, ok := <-depositActionsCh:
			if !ok {
				return fmt.Errorf("deposit action channel closed")
			}
			log.Printf("new deposit action: %v", string(action.TrxID))
			if err := s.publishAction(action); err != nil {
				log.Printf("Publish deposit action failed: %v", err)
				continue
			}
			s.depositLastBlockNum = action.BlockNum
		case action, ok := <-withdrawActionsCh:
			if !ok {
				return fmt.Errorf("withdraw action channel closed")
			}
			log.Printf("new withdraw action: %v", string(action.TrxID))
			if err := s.publishAction(action); err != nil {
				log.Printf("Publish withdraw action failed: %v", err)
				continue
			}
			s.withdrawLastBlockNum = action.BlockNum

		case action, ok := <-accountActionsCh:
			if !ok {
				return fmt.Errorf("account action channel closed")
			}
			log.Printf("new updateauth  action: %v", string(action.TrxID))
			if err := s.publishAction(action); err != nil {
				log.Printf("Publish account updateauth action failed: %v", err)
				continue
			}
			s.accountLastBlockNum = action.BlockNum
		case action, ok := <-poolActionsCh:
			if !ok {
				return fmt.Errorf("pool action channel closed")
			}
			log.Printf("new pool action: %v", string(action.TrxID))
			if err := s.publishAction(action); err != nil {
				log.Printf("Publish pool action failed: %v", err)
				continue
			}
			s.poolLastBlockNum = action.BlockNum
		}
	}
}

func (s *Service) Stop() error {
	s.publisher.Stop()
	return s.streamClient.Close()
}

func (s *Service) publishAction(action hyperion.Action) error {
	return s.publisher.Publish(TopicActionSync, action)
}
