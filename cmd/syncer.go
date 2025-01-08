package cmd

import (
	"context"
	"exapp-go/config"
	"exapp-go/internal/db/db"
	"exapp-go/internal/services/syncer"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/urfave/cli/v2"
)

var SyncCmd = &cli.Command{
	Name:  "syncer",
	Usage: "sync data from hyperion",
	Action: func(c *cli.Context) error {
		err := config.Load(c.String("config"))
		if err != nil {
			log.Printf("load config err: %v\n", err)
			return err
		}

		db.New()

		cfg := config.Conf()
		srv, err := syncer.NewService(cfg.Hyperion, cfg.Nsq, cfg.Cdex)
		if err != nil {
			log.Printf("create syncer service failed: %v\n", err)
			return err
		}

		ctx, cancel := context.WithCancel(context.Background())
		go func() {
			log.Println("syncer service start ...")
			if err := srv.Start(ctx); err != nil {
				log.Printf("syncer service start failed: %v\n", err)
			}
		}()

		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

		<-quit
		cancel()
		log.Println("stop syncer service ...")

		if err := srv.Stop(); err != nil {
			log.Printf("syncer service stop failed: %v\n", err)
			return err
		}

		log.Println("syncer service stop success")

		return nil
	},
}
