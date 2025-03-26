package db

import (
	"context"
	"exapp-go/config"
	"exapp-go/pkg/queryparams"
	"exapp-go/pkg/utils"
	"fmt"
	"testing"
	"time"

	"github.com/shopspring/decimal"
)

func TestRepo_UpdateDepth(t *testing.T) {
	utils.WorkInProjectPath("exapp-go")
	config.Load("config/config_dev.yaml")
	r := New()
	changes, err := r.UpdateDepth(context.Background(), []UpdateDepthParams{
		{PoolID: 1, IsBuy: false, Price: decimal.NewFromFloat(100.0), Amount: decimal.NewFromInt(-20), UniqID: "1"},
	})
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(changes)
	depth, err := r.GetDepth(context.Background(), 1)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(depth)

}

func TestQueryTransactionsRecord(t *testing.T) {
	utils.WorkInProjectPath("exapp-go")
	config.Load("config/config.yaml")

	params := &queryparams.QueryParams{
		Offset: 0,
		Limit:  10,
		CustomQuery: map[string][]interface{}{
			"symbol":     []any{"USDT"},
			"start_time": []any{time.Now().Add(-time.Hour * 24 * 30)},
			"end_time":   []any{time.Now()},
		},
	}

	r := New()
	record, totle, err := r.QueryTransactionsRecord(context.Background(), params)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(record, totle)
}

func TestGetDepositAmountTotal(t *testing.T) {
	utils.WorkInProjectPath("exapp-go")
	config.Load("config/config.yaml")

	r := New()
	record, err := r.GetDepositAmountTotal(context.Background(), "2024-10-01 00:00:00", "2025-04-01 23:59:59")
	if err != nil {
		t.Fatal(err)
	}

	for _, v := range record {
		fmt.Println(v.Symbol, v.Amount)
	}

	fmt.Println(len(record))
}

func TestGetWithdrawAmountTotal(t *testing.T) {
	utils.WorkInProjectPath("exapp-go")
	config.Load("config/config.yaml")

	r := New()
	record, err := r.GetWithdrawAmountTotal(context.Background(), "2024-10-01 00:00:00", "2025-04-01 23:59:59")
	if err != nil {
		t.Fatal(err)
	}

	for _, v := range record {
		fmt.Println(v.Symbol, v.Amount)
	}

	fmt.Println(len(record))
}
