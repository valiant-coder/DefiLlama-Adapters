package cmd

import (
	"exapp-go/config"
	"exapp-go/internal/services/points"
	"github.com/urfave/cli/v2"
	"log"
)

var UserPointsCmd = &cli.Command{
	Name:  "points",
	Usage: "user points service",
	Action: func(c *cli.Context) error {
		err := config.Load(c.String("config"))
		if err != nil {
			log.Printf("load config err: %v\n", err)
			return err
		}
		
		if err := points.Start(); err != nil {
			
			log.Printf("start user points service failed: %v\n", err)
			return err
		}
		
		return nil
	},
}
