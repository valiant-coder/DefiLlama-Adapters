package db

import (
	"context"
	"exapp-go/config"
	"exapp-go/pkg/utils"
	"fmt"
	"testing"

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
