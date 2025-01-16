package cron

import (
	"context"
	"exapp-go/config"
	"exapp-go/internal/db/ckhdb"
	"exapp-go/internal/db/db"
	"exapp-go/pkg/eos"
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
	log.Println("begin sync pool stats...")
	ctx := context.Background()
	err := s.ckhdb.UpdatePoolStats(ctx)
	if err != nil {
		log.Println("failed to update pool stats", err)
	}
	err = s.ckhdb.OptimizePoolStats(ctx)
	if err != nil {
		log.Println("failed to optimize pool stats", err)
	}
	log.Println("sync pool stats done")
}

func (s *Service) PowerUp() {
	log.Println("begin powerup for payer account...")
	ctx := context.Background()
	conf := config.Conf().Eos

	if !conf.PowerUp.Enabled {
		log.Println("powerup is disabled, skipping...")
		return
	}

	err := eos.PowerUp(
		ctx,
		conf.NodeURL,
		conf.PayerAccount,
		conf.PayerPrivateKey,
		conf.PowerUp.NetEOS,
		conf.PowerUp.CPUEOS,
		conf.PowerUp.MaxPayment,
	)
	if err != nil {
		log.Printf("failed to powerup for payer account: %v\n", err)
		return
	}

	err = eos.PowerUp(
		ctx,
		conf.NodeURL,
		conf.Exapp.Actor,
		conf.Exapp.ActorPrivateKey,
		conf.PowerUp.NetEOS,
		conf.PowerUp.CPUEOS,
		conf.PowerUp.MaxPayment,
	)
	if err != nil {
		log.Printf("failed to powerup for exapp actor: %v\n", err)
		return
	}

	log.Println("powerup for payer account done")
}

func (s *Service) Run() error {
	c := cron.New()

	addSyncFuncs(c, "0 * * * * *", s.SyncPoolStats)

	addSyncFuncs(c, "0 0 1 * * *", s.PowerUp)

	c.Run()
	return nil
}
