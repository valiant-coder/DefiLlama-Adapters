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

	orders, err := ckhRepo.GetOrdersCoinQuantity(context.Background(), "2024-10-01", "2025-10-31")
	if err != nil {
		t.Error(err)
	}

	for _, order := range orders {
		fmt.Println(order.PoolBaseCoin, order.ExecutedQuantity)
	}
}

func TestGetOrdersSymbolTotal(t *testing.T) {

	utils.WorkInProjectPath("exapp-go")
	config.Load("config/config.yaml")
	ckhRepo := New()

	orders, err := ckhRepo.GetOrdersSymbolQuantity(context.Background(), "2024-10-01", "2025-10-31")
	if err != nil {
		t.Error(err)
	}

	for _, order := range orders {
		fmt.Println(order.Symbol, order.Quantity, order.Price)
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

	for _, order := range orders {
		fmt.Println(*order)
	}
	fmt.Println(total)
}

func TestGetOrdersCoinFee(t *testing.T) {

	utils.WorkInProjectPath("exapp-go")
	config.Load("config/config.yaml")
	ckhRepo := New()

	orders, err := ckhRepo.GetOrdersCoinFee(context.Background(), "2024-10-01", "2025-10-31")
	if err != nil {
		t.Error(err)
	}

	for _, order := range orders {
		fmt.Println(order.PoolBaseCoin, order.Fee)
	}
}

func TestGetOrdersSymbolFee(t *testing.T) {

	utils.WorkInProjectPath("exapp-go")
	config.Load("config/config.yaml")
	ckhRepo := New()

	orders, err := ckhRepo.GetOrdersSymbolFee(context.Background(), "2024-10-01", "2025-10-31")
	if err != nil {
		t.Error(err)
	}

	for _, order := range orders {
		fmt.Println(order.Symbol, order.MakerFee, order.TakerFee)
	}
}
