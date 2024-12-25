package cmd

import (
	"context"
	"exapp-go/config"
	"exapp-go/internal/db/db"
	"exapp-go/internal/services/handler"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/urfave/cli/v2"
)

var HandlerCmd = &cli.Command{
	Name:  "handler",
	Usage: "handle hyperion data",
	Action: func(c *cli.Context) error {
		err := config.Load(c.String("config"))
		if err != nil {
			log.Printf("load config err: %v\n", err)
			return err
		}

		db.New()

		srv, err := handler.NewService()
		if err != nil {
			log.Printf("create handler service failed: %v\n", err)
			return err
		}

		ctx, cancel := context.WithCancel(context.Background())
		go func() {
			if err := srv.Start(ctx); err != nil {
				log.Printf("handler service start failed: %v\n", err)
			}
		}()

		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

		<-quit
		cancel()
		log.Println("stop handler service ...")

		if err := srv.Stop(ctx); err != nil {
			log.Printf("handler service stop failed: %v\n", err)
			return err
		}

		log.Println("handler service stop success")

		return nil
	},
}
