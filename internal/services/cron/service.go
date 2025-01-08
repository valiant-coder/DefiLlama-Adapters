package cron

import (
	"context"
	"exapp-go/internal/db/ckhdb"
	"exapp-go/internal/db/db"
	"log"

	"github.com/robfig/cron"
)

type Service struct {
	repo  *db.Repo
	ckhdb *ckhdb.ClickHouseRepo
}

func NewService() *Service {
	return &Service{
		repo:  db.New(),
		ckhdb: ckhdb.New(),
	}
}

func addSyncFuncs(c *cron.Cron, spec string, cmdList ...func()) {
	for _, cmd := range cmdList {
		c.AddFunc(spec, cmd)
	}
}

func (s *Service) SyncPoolStats() {
	ctx := context.Background()
	err := s.ckhdb.UpdatePoolStats(ctx)
	if err != nil {
		log.Println("failed to update pool stats", err)
	}
	err = s.ckhdb.OptimizePoolStats(ctx)
	if err != nil {
		log.Println("failed to optimize pool stats", err)
	}

}

func (s *Service) Run() error {
	c := cron.New()

	addSyncFuncs(c, "0 0 * * *", s.SyncPoolStats)

	c.Run()
	return nil
}
