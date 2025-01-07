package main

import (
	"exapp-go/cmd"
	"log"
	"os"

	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:  "exapp",
		Usage: "exapp go",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "config",
				Aliases: []string{"c"},
				Value:   "config/config.toml",
				Usage:   "config file path",
			},
			&cli.StringFlag{
				Name:  "log-dir",
				Value: "logs",
				Usage: "log directory",
			},
		},
		Commands: []*cli.Command{
			cmd.MarketplaceApi,
			cmd.SyncCmd,
			cmd.HandlerCmd,
			cmd.CronCmd,
		},
	}
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}

}
