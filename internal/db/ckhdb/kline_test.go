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
	testTime := time.Now()
	kline, err := repo.GetLastKlineBefore(context.Background(), 1, "1m", testTime)
	if err != nil {
		t.Fatalf("failed to get last kline before: %v", err)
	}
	if kline == nil {
		t.Fatal("expected kline to be non-nil")
	}
	t.Logf("last kline before: %+v", kline)
}

func TestClickHouseRepo_GetLatestKlines(t *testing.T) {
	utils.WorkInProjectPath("exapp-go")
	config.Load("config/config_dev.yaml")
	repo := New()
	klines, err := repo.GetLatestKlines(context.Background(), 1)
	if err != nil {
		t.Fatalf("failed to get latest klines: %v", err)
	}
	t.Logf("latest klines: %+v", klines)
}
