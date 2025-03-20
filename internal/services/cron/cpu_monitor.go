package cron

import (
	"context"
	"exapp-go/config"
	"exapp-go/pkg/eos"
	"log"

	eosapi "github.com/eoscanada/eos-go"
)

// MonitorCPUResources checks the CPU resources of the payer account and performs powerup if needed
func (s *Service) MonitorCPUResources() {
	log.Println("begin monitoring CPU resources for payer account...")
	ctx := context.Background()
	conf := config.Conf().Eos

	if !conf.PowerUp.Enabled || !conf.PowerUp.CPUMonitorEnabled {
		log.Println("CPU monitoring is disabled, skipping...")
		return
	}

	// Get account info to check CPU resources
	api := eosapi.New(conf.NodeURL)
	account, err := api.GetAccount(ctx, eosapi.AccountName(conf.PayerAccount))
	if err != nil {
		log.Printf("failed to get account info: %v\n", err)
		return
	}

	// Calculate CPU usage percentage
	cpuUsed := float64(account.CPULimit.Used)
	cpuMax := float64(account.CPULimit.Max)
	if cpuMax == 0 {
		log.Println("invalid CPU limit (max is 0)")
		return
	}

	cpuUsagePercent := (cpuUsed / cpuMax) * 100
	availablePercent := 100 - cpuUsagePercent

	log.Printf("CPU usage: %.2f%%, available: %.2f%%\n", cpuUsagePercent, availablePercent)

	// If available CPU is below threshold, trigger powerup
	if availablePercent < conf.PowerUp.CPUThreshold {
		log.Printf("CPU available (%.2f%%) is below threshold (%.2f%%), triggering powerup...\n",
			availablePercent, conf.PowerUp.CPUThreshold)

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

		log.Println("powerup for payer account completed successfully")
	} else {
		log.Println("CPU resources are sufficient, no powerup needed")
	}
}
