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
	TopicActionSync   = "dex_action_sync"
	ChannelActionSync = "dex_action_sync"
)

type Service struct {
	nsqCfg    config.NsqConfig
	ckhRepo   *ckhdb.ClickHouseRepo
	repo      *db.Repo
	worker    *nsqutil.Worker
	poolCache map[uint64]*ckhdb.Pool
}

func NewService() (*Service, error) {
	ckhRepo := ckhdb.New()
	repo := db.New()

	return &Service{
		ckhRepo: ckhRepo,
		repo:    repo,
	}, nil
}

func (s *Service) Start(ctx context.Context) error {
	worker := nsqutil.NewWorker(ChannelActionSync, s.nsqCfg.Lookupd, s.nsqCfg.LookupTTl)
	s.worker = worker
	s.worker.Consume(TopicActionSync, s.HandleMessage)
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
	switch action.Act.Name {
	case "emitplaced":
		return s.handleCreateOrder(action)
	case "emitcanceled":
		return s.handleCancelOrder(action)
	case "emitfilled":
		return s.handleMatchOrder(action)
	default:
		log.Printf("Unknown action: %s", action.Act.Name)
		return nil
	}
}
