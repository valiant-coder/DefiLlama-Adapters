package ckhdb

import (
	"context"
	"exapp-go/config"
	"exapp-go/pkg/utils"
	"testing"
)

func TestClickHouseRepo_UpdatePoolStats(t *testing.T) {
	utils.WorkInProjectPath("exapp-go")
	config.Load("config/config_dev.yaml")
	ckhRepo := New()
	err := ckhRepo.UpdatePoolStats(context.Background())
	if err != nil {
		t.Errorf("ClickHouseRepo.UpdatePoolStats() error = %v", err)
	}
}
