package dex

import (
	"context"
	"testing"
)

func TestNewClient(t *testing.T) {
	// jungle4
	nodeUrl := "https://jungle4.cryptolions.io"
	dexContract := "mightyjungle"
	poolContract := "colorfuljung"
	client := NewClient(nodeUrl, dexContract, poolContract)
	pools, err := client.GetPools(context.Background())
	if err != nil {
		t.Fatalf("failed to get pools: %v", err)
	}
	t.Logf("pools: %+v", pools)
	orders, err := client.GetOrders(context.Background(), 0, true)
	if err != nil {
		t.Fatalf("failed to get orders: %v", err)
	}
	t.Logf("orders: %+v", orders)
	userFunds, err := client.GetUserFunds(context.Background(), "playfullion4")
	if err != nil {
		t.Fatalf("failed to get user funds: %v", err)
	}
	t.Logf("user funds: %+v", userFunds)
}
