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
	cdexCfg   config.CdexConfig
	eosCfg    config.EosConfig
	hyperionCfg config.HyperionConfig
}

func NewService() (*Service, error) {
	ckhRepo := ckhdb.New()
	repo := db.New()
	cfg := config.Conf()

	return &Service{
		ckhRepo: ckhRepo,
		repo:    repo,
		nsqCfg:  cfg.Nsq,
		poolCache: make(map[uint64]*db.Pool),
		cdexCfg: cfg.Cdex,
		eosCfg:  cfg.Eos,
		hyperionCfg: cfg.Hyperion,
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
	return nil
}

func (s *Service) HandleMessage(msg *nsq.Message) error {
	var action hyperion.Action
	if err := json.Unmarshal(msg.Body, &action); err != nil {
		log.Printf("Unmarshal action failed: %v", err)
		return nil
	}
	if action.Act.Account != s.cdexCfg.EventContract && action.Act.Account != s.cdexCfg.PoolContract {
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
	default:
		log.Printf("Unknown action: %s", action.Act.Name)
		return nil
	}
}
