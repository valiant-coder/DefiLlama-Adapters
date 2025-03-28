package ckhdb

import (
	"context"
	"exapp-go/config"
	"exapp-go/pkg/queryparams"
	"exapp-go/pkg/utils"
	"fmt"
	"testing"
)

func TestGetOrdersCoinTotal(t *testing.T) {

	utils.WorkInProjectPath("exapp-go")
	config.Load("config/config.yaml")
	ckhRepo := New()

	orders, err := ckhRepo.GetOrdersCoinTotal(context.Background(), "2024-10-01", "2025-10-31")
	if err != nil {
		t.Error(err)
	}

	for _, ororder := range orders {
		fmt.Println(ororder.PoolBaseCoin, ororder.ExecutedQuantity)
	}
}

func TestGetOrdersSymbolTotal(t *testing.T) {

	utils.WorkInProjectPath("exapp-go")
	config.Load("config/config.yaml")
	ckhRepo := New()

	orders, err := ckhRepo.GetOrdersSymbolTotal(context.Background(), "2024-10-01", "2025-10-31")
	if err != nil {
		t.Error(err)
	}

	for _, ororder := range orders {
		fmt.Println(ororder.PoolSymbol, ororder.ExecutedQuantity)
	}
}

func TestQueryHistoryOrdersList(t *testing.T) {

	utils.WorkInProjectPath("exapp-go")
	config.Load("config/config.yaml")
	ckhRepo := New()

	params := &queryparams.QueryParams{
		Offset: 0,
		Limit:  10,
		CustomQuery: map[string][]any{
			"app":            {"app.1dex"},
			"pool_base_coin": {"goldgoldgold-GOLD"},
			"trader":         {"pablot2n1113"},
			"pool_symbol":    {"goldgoldgold-GOLD-usd1usd1usd1-USDI"},
		},
	}

	orders, total, err := ckhRepo.QueryHistoryOrdersList(context.Background(), params)
	if err != nil {
		t.Error(err)
	}

	for _, ororder := range orders {
		fmt.Println(*ororder)
	}
	fmt.Println(total)
}
