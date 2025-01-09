package marketplace

import (
	"context"
	"exapp-go/config"
	"exapp-go/pkg/utils"
	"log"
	"testing"
)

func TestTradeService_GetLatestTrades(t *testing.T) {
	utils.WorkInProjectPath("exapp-go")
	config.Load("config/config_dev.yaml")
	tradeService := NewTradeService()
	trades, err := tradeService.GetLatestTrades(context.Background(), 0, 10)
	if err != nil {
		t.Errorf("TradeService.GetLatestTrades() error = %v", err)
	}
	log.Printf("TradeService.GetLatestTrades() = %v", trades)
}
