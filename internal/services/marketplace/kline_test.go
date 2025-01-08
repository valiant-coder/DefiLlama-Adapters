package marketplace

import (
	"context"
	"exapp-go/config"
	"exapp-go/pkg/utils"
	"log"
	"testing"
	"time"
)

func TestKlineService_GetKline(t *testing.T) {
	utils.WorkInProjectPath("exapp-go")
	config.Load("config/config_dev.yaml")
	klineService := NewKlineService()

	klines, err := klineService.GetKline(context.Background(), 0, "1m", time.Now().Add(-time.Hour*8), time.Now())
	if err != nil {
		t.Errorf("KlineService.GetKline() error = %v", err)
	}
	for _, kline := range klines {
		log.Printf("KlineService.GetKline() = %v", kline.Timestamp)
	}
}
