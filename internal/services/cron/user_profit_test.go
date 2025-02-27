package cron

import (
	"exapp-go/config"
	"exapp-go/pkg/utils"
	"testing"
)

func TestService_HandleUserProfit(t *testing.T) {
	utils.WorkInProjectPath("exapp-go")
	config.Load("config/config_dev.yaml")
	service := NewService()
	service.HandleUserProfit()
}
