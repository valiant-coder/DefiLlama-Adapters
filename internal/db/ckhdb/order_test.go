package ckhdb

import (
	"context"
	"exapp-go/config"
	"exapp-go/pkg/queryparams"
	"exapp-go/pkg/utils"
	"fmt"
	"testing"
	"time"
)

func TestGetOrdersCoinTotal(t *testing.T) {

	utils.WorkInProjectPath("exapp-go")
	config.Load("config/config.yaml")
	ckhRepo := New()

	startTime, _ := time.Parse("2006-01-02", "2024-10-01")
	endTime, _ := time.Parse("2006-01-02", "2025-10-31")

	total, err := ckhRepo.GetOrdersCoinTotal(context.Background(), startTime, endTime)
	if err != nil {
		t.Error(err)
	}

	fmt.Println(total)
}

func TestGetOrdersCoinQuantity(t *testing.T) {

	utils.WorkInProjectPath("exapp-go")
	config.Load("config/config.yaml")
	ckhRepo := New()

	startTime, _ := time.Parse("2006-01-02", "2024-10-01")
	endTime, _ := time.Parse("2006-01-02", "2025-10-31")

	orders, err := ckhRepo.GetOrdersCoinQuantity(context.Background(), startTime, endTime)
	if err != nil {
		t.Error(err)
	}

	for _, order := range orders {
		fmt.Println(order.PoolBaseCoin, order.ExecutedQuantity)
	}
}

func TestGetOrdersSymbolQuantity(t *testing.T) {

	utils.WorkInProjectPath("exapp-go")
	config.Load("config/config.yaml")
	ckhRepo := New()

	startTime, _ := time.Parse("2006-01-02", "2024-10-01")
	endTime, _ := time.Parse("2006-01-02", "2025-10-31")

	orders, err := ckhRepo.GetOrdersSymbolQuantity(context.Background(), startTime, endTime)
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

func TestGetOrdersFeeTotal(t *testing.T) {
	utils.WorkInProjectPath("exapp-go")
	config.Load("config/config.yaml")
	ckhRepo := New()

	startTime, _ := time.Parse("2006-01-02", "2024-10-01")
	endTime, _ := time.Parse("2006-01-02", "2025-10-31")

	total, err := ckhRepo.GetOrdersFeeTotal(context.Background(), startTime, endTime)
	if err != nil {
		t.Error(err)
	}

	fmt.Println(total)
}

func TestGetOrdersCoinFee(t *testing.T) {

	utils.WorkInProjectPath("exapp-go")
	config.Load("config/config.yaml")
	ckhRepo := New()

	startTime, _ := time.Parse("2006-01-02", "2024-10-01")
	endTime, _ := time.Parse("2006-01-02", "2025-10-31")

	orders, err := ckhRepo.GetOrdersCoinFee(context.Background(), startTime, endTime)
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

	startTime, _ := time.Parse("2006-01-02", "2024-10-01")
	endTime, _ := time.Parse("2006-01-02", "2025-10-31")

	orders, err := ckhRepo.GetOrdersSymbolFee(context.Background(), startTime, endTime)
	if err != nil {
		t.Error(err)
	}

	for _, order := range orders {
		fmt.Println(order.Symbol, order.MakerFee, order.TakerFee)
	}
}
