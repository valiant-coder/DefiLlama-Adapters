package cmd

import (
	"exapp-go/api/marketplace"
	"exapp-go/config"
	"exapp-go/internal/db/db"
	"exapp-go/pkg/trace"
	"fmt"

	"github.com/urfave/cli/v2"
)

var MarketplaceApi = &cli.Command{
	Name:  "marketplace-api",
	Usage: "cdex go marketplace api",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "addr",
			Usage: "http service address",
			Value: ":8080",
		},
		&cli.BoolFlag{
			Name:  "release",
			Value: false,
			Usage: "release mode",
		},
	},
	Action: func(c *cli.Context) error {
		err := config.Load(c.String("config"))
		if err != nil {
			fmt.Printf("load config err: %v\n", err)
			return err
		}
		_ = db.New()
		err = trace.Init("api-marketplace")
		if err != nil {
			fmt.Printf("trace init err: %v\n", err)
			return err
		}
		return marketplace.Run(c.String("addr"), c.Bool("release"))
	},
}
