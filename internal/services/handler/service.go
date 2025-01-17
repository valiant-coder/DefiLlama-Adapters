package handler

import (
	"context"
	"encoding/json"
	"exapp-go/config"
	"exapp-go/internal/db/ckhdb"
	"exapp-go/internal/db/db"
	"exapp-go/pkg/hyperion"
	"exapp-go/pkg/nsqutil"
	"log"

	"github.com/nsqio/go-nsq"
)

const (
	TopicActionSync   = "cdex_action_sync"
	ChannelActionSync = "channel_cdex_action_sync"
)

type Service struct {
	nsqCfg    config.NsqConfig
	ckhRepo   *ckhdb.ClickHouseRepo
	repo      *db.Repo
	worker    *nsqutil.Worker
	poolCache map[uint64]*db.Pool
	eosCfg    config.EosConfig
	cdexCfg   config.CdexConfig
	exappCfg  config.ExappConfig
	publisher *NSQPublisher
}

func NewService() (*Service, error) {
	ckhRepo := ckhdb.New()
	repo := db.New()
	cfg := config.Conf()

	publisher, err := NewNSQPublisher(cfg.Nsq.Nsqds)
	if err != nil {
		return nil, err
	}

	return &Service{
		ckhRepo:   ckhRepo,
		repo:      repo,
		nsqCfg:    cfg.Nsq,
		poolCache: make(map[uint64]*db.Pool),
		eosCfg:    cfg.Eos,
		cdexCfg:   cfg.Eos.CdexConfig,
		exappCfg:  cfg.Eos.Exapp,
		publisher: publisher,
	}, nil
}

func (s *Service) Start(ctx context.Context) error {
	worker := nsqutil.NewWorker(ChannelActionSync, s.nsqCfg.Lookupd, s.nsqCfg.LookupTTl)
	s.worker = worker
	err := s.worker.Consume(TopicActionSync, s.HandleMessage)
	if err != nil {
		log.Printf("Consume action sync failed: %v", err)
		return err
	}
	<-ctx.Done()
	return nil
}

func (s *Service) Stop(ctx context.Context) error {
	s.worker.StopConsume()
	if s.publisher != nil {
		s.publisher.Close()
	}
	return nil
}

func (s *Service) HandleMessage(msg *nsq.Message) error {
	log.Println("get new action")
	var action hyperion.Action
	if err := json.Unmarshal(msg.Body, &action); err != nil {
		log.Printf("Unmarshal action failed: %v", err)
		return nil
	}
	if action.Act.Account != s.cdexCfg.EventContract && action.Act.Account != s.cdexCfg.PoolContract && action.Act.Account != s.exappCfg.AssetContract {
		return nil
	}
	switch action.Act.Name {
	case "emitplaced":
		return s.handleCreateOrder(action)
	case "emitcanceled":
		return s.handleCancelOrder(action)
	case "emitfilled":
		return s.handleMatchOrder(action)
	case "create":
		return s.handleCreatePool(action)
	case "lognewacc":
		return s.handleNewAccount(action)
	case "logdeposit":
		return s.handleDeposit(action)
	case "logwithdraw":
		return s.handleWithdraw(action)
	default:
		log.Printf("Unknown action: %s", action.Act.Name)
		return nil
	}
}
