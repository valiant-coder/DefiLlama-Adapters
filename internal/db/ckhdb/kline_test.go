package ckhdb

import (
	"context"
	"exapp-go/config"
	"exapp-go/pkg/utils"
	"testing"
	"time"
)

func TestClickHouseRepo_GetLastKlineBefore(t *testing.T) {
	utils.WorkInProjectPath("exapp-go")
	config.Load("config/config_dev.yaml")
	repo := New()
	kline, err := repo.GetLastKlineBefore(context.Background(), 1, "1m", time.Now())
	if err != nil {
		t.Errorf("failed to get last kline before: %v", err)
	}
	t.Logf("last kline before: %+v", kline)
}
