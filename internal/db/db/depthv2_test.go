package db

import (
	"context"
	"exapp-go/config"
	"exapp-go/pkg/utils"
	"fmt"
	"sort"
	"testing"

	"github.com/shopspring/decimal"
)

func TestDepthV2(t *testing.T) {
	utils.WorkInProjectPath("exapp-go")
	config.Load("config/config_dev.yaml")
	r := New()

	// Initialize buy side depth manager
	ctx := context.Background()

	// Add orders (automatically update all precision levels)
	err := r.UpdateDepthV2(ctx, "BTCUSDT", "buy", decimal.NewFromFloat(0.1234), decimal.NewFromInt(100)) // Affects 0.001, 0.01, 0.1 precision levels
	if err != nil {
		fmt.Println(err)
		t.Fatal(err)
	}
	err = r.UpdateDepthV2(ctx, "BTCUSDT", "buy", decimal.NewFromFloat(0.1245), decimal.NewFromInt(200))
	if err != nil {
		fmt.Println(err)
		t.Fatal(err)
	}
	err = r.UpdateDepthV2(ctx, "BTCUSDT", "buy", decimal.NewFromFloat(1.2345), decimal.NewFromInt(300))
	if err != nil {
		fmt.Println(err)
		t.Fatal(err)
	}

	// Query depth at different precision levels
	fmt.Println("Depth at precision 0.001:")
	d, _ := r.GetDepthV2(ctx, "BTCUSDT", "buy", "0.001", 10)
	printDepth(d)

	fmt.Println("\nDepth at precision 0.01:")
	d, _ = r.GetDepthV2(ctx, "BTCUSDT", "buy", "0.01", 10)
	printDepth(d)

	fmt.Println("\nDepth at precision 0.1:")
	d, _ = r.GetDepthV2(ctx, "BTCUSDT", "buy", "0.1", 10)
	printDepth(d)
}

// Helper function: print depth data
func printDepth(depth map[decimal.Decimal]decimal.Decimal) {
	var prices []decimal.Decimal
	for p := range depth {
		prices = append(prices, p)
	}

	// Sort buy side by descending price
	sort.Slice(prices, func(i, j int) bool {
		return prices[i].GreaterThan(prices[j])
	})

	for _, p := range prices {
		fmt.Printf("Price: %s  Quantity: %s\n", p.String(), depth[p].String())
	}
}
