package cron

import (
	"context"
	"exapp-go/config"
	"exapp-go/internal/db/ckhdb"
	"exapp-go/internal/db/db"
	"exapp-go/internal/entity"
	"exapp-go/pkg/eos"
	"exapp-go/pkg/nsqutil"
	"log"

	"github.com/robfig/cron"
)

type Service struct {
	repo         *db.Repo
	ckhdb        *ckhdb.ClickHouseRepo
	nsqPublisher *nsqutil.Publisher
}

func NewService() *Service {
	nsqCfg := config.Conf().Nsq
	return &Service{
		repo:         db.New(),
		ckhdb:        ckhdb.New(),
		nsqPublisher: nsqutil.NewPublisher(nsqCfg.Nsqds),
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
		conf.OneDex.Actor,
		conf.OneDex.ActorPrivateKey,
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

func (s *Service) SyncAndBroadcastPoolStats() {
	log.Println("begin sync and broadcast pool stats...")
	ctx := context.Background()

	stats, err := s.ckhdb.ListPoolStats(ctx)
	if err != nil {
		log.Printf("failed to list pool stats: %v\n", err)
		return
	}
	for _, stat := range stats {

		poolStats := entity.PoolStatusFromDB(stat)

		msg := struct {
			Type string      `json:"type"`
			Data interface{} `json:"data"`
		}{
			Type: "pool_stats_update",
			Data: poolStats,
		}

		err = s.nsqPublisher.Publish("cdex_updates", msg)
		if err != nil {
			log.Printf("failed to publish pool stats message: %v\n", err)
			continue
		}
	}

	log.Println("sync and broadcast pool stats done")
}

func (s *Service) Run() error {
	c := cron.New()

	addSyncFuncs(c, "*/2 * * * * *", s.SyncPoolStats)
	addSyncFuncs(c, "*/2 * * * * *", s.SyncAndBroadcastPoolStats)
	addSyncFuncs(c, "0 0 1 * * *", s.PowerUp)
	addSyncFuncs(c, "0 0 * * * *", s.HandleUserProfit)

	c.Run()
	return nil
}
