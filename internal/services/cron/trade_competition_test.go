package cron

import (
	"exapp-go/config"
	"exapp-go/pkg/utils"
	"fmt"
	"testing"
)

func TestRecordUserBalances(t *testing.T) {

	utils.WorkInProjectPath("exapp-go")
	config.Load("config/config.yaml")
	service := NewService()

	err := service.recordUserBalances()
	if err != nil {
		t.Errorf("recordUserBalances error %v", err)
	}

	fmt.Println("success")
}
