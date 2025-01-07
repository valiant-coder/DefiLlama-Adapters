package cmd

import (
	"exapp-go/config"
	"exapp-go/internal/db/ckhdb"
	"exapp-go/internal/db/db"
	"exapp-go/internal/services/cron"
	"log"

	"github.com/urfave/cli/v2"
)

var CronCmd = &cli.Command{
	Name:  "cron",
	Usage: "cron job",
	Action: func(c *cli.Context) error {
		err := config.Load(c.String("config"))
		if err != nil {
			log.Printf("load config err: %v\n", err)
			return err
		}

		db.New()
		ckhdb.New()

		service := cron.NewService()
		return service.Run()
	},
}
