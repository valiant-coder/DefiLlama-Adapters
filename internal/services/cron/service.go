package cron

import (
	"context"
	"exapp-go/config"
	"exapp-go/internal/db/ckhdb"
	"exapp-go/internal/db/db"
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

func (s *Service) Run() error {
	c := cron.New()

	addSyncFuncs(c, "*/2 * * * * *", s.SyncPoolStats)
	addSyncFuncs(c, "*/2 * * * * *", s.SyncAndBroadcastPoolStats)
	addSyncFuncs(c, "0 0 1 * * *", s.PowerUp)
	addSyncFuncs(c, "0 0 * * * *", s.HandleTradeCompetition)

	c.Run()
	return nil
}
