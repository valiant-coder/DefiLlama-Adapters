package db

import (
	"context"
	"exapp-go/config"
	"exapp-go/pkg/utils"
	"fmt"
	"testing"

	"github.com/shopspring/decimal"
)

func TestDepthV2(t *testing.T) {
	utils.WorkInProjectPath("exapp-go")
	config.Load("config/config_dev.yaml")
	r := New()

	ctx := context.Background()
	poolID := uint64(10)

	testCases := []UpdateDepthParams{
		{
			PoolID: poolID,
			IsBuy:  true,
			Price:  decimal.NewFromFloat(0.1234),
			Amount: decimal.NewFromInt(100),
		},
		{
			PoolID: poolID,
			IsBuy:  true,
			Price:  decimal.NewFromFloat(0.1245),
			Amount: decimal.NewFromInt(200),
		},
		{
			PoolID: poolID,
			IsBuy:  true,
			Price:  decimal.NewFromFloat(1.2345),
			Amount: decimal.NewFromInt(300),
		},
		{
			PoolID: poolID,
			IsBuy:  true,
			Price:  decimal.NewFromFloat(0.1234),
			Amount: decimal.NewFromFloat(-80.1),
		},
		{
			PoolID: poolID,
			IsBuy:  false,
			Price:  decimal.NewFromFloat(2.3456),
			Amount: decimal.NewFromFloat(100),
		},
		{
			PoolID: poolID,
			IsBuy:  false,
			Price:  decimal.NewFromFloat(1.00001),
			Amount: decimal.NewFromFloat(80.1),
		},
		{
			PoolID: poolID,
			IsBuy:  false,
			Price:  decimal.NewFromFloat(1.00001),
			Amount: decimal.NewFromFloat(-20),
		},
	}

	changes, err := r.UpdateDepthV2(ctx, testCases)
	if err != nil {
		t.Fatalf("Failed to update depth: %v", err)
	}
	if len(changes) != len(testCases) {
		t.Errorf("Expected %d changes, got %d", len(testCases), len(changes))
	}

	precisions := []string{"0.001", "0.01", "0.1"}
	for _, precision := range precisions {
		fmt.Printf("\nDepth at precision %s:\n", precision)
		depth, err := r.GetDepthV2(ctx, poolID, precision, 10)
		if err != nil {
			t.Fatalf("Failed to get depth at precision %s: %v", precision, err)
		}
		printDepth(depth)
	}
	r.ClearDepthsV2(ctx, poolID)
}

// Helper function: print depth data
func printDepth(depth Depth) {
	fmt.Printf("Pool ID: %d\n", depth.PoolID)

	fmt.Println("Bids:")
	for _, bid := range depth.Bids {
		fmt.Printf("Price: %s  Quantity: %s\n", bid[0], bid[1])
	}

	fmt.Println("Asks:")
	for _, ask := range depth.Asks {
		fmt.Printf("Price: %s  Quantity: %s\n", ask[0], ask[1])
	}
}
