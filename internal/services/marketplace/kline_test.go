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

	start := time.Unix(1736649481, 0)
	end := time.Unix(1736822281, 0)

	klines, err := klineService.GetKline(context.Background(), 1, "4h", start, end)
	if err != nil {
		t.Errorf("KlineService.GetKline() error = %v", err)
	}
	for _, kline := range klines {
		log.Printf("KlineService.GetKline() = %v", kline.Timestamp)
	}
}
