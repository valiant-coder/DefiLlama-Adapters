package onedex

import (
	"context"
	"testing"

	"github.com/eoscanada/eos-go"
)

func TestGetSubaccountBalances(t *testing.T) {

	client := eos.New("http://44.223.68.11:8888")
	balances, err := GetSubaccountBalances(context.Background(), client, "makerfundacc", "subacc1")
	if err != nil {
		t.Fatalf("Failed to get subaccount balances: %v", err)
	}
	t.Logf("Subaccount balances: %v", balances)

}
