package eth

import (
	"context"
	"fmt"
	"testing"
)

func TestNewEthScanClient(t *testing.T) {
	client := NewEthScanClient("https://scan2.exactsat.io", "1234567890")

	balances, err := client.GetTokenBalancesByAddress(context.Background(), "0xEc75C4a6c4E048AAAD7be187c1631902dD7349F1")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("balances: %+v\n", balances)
}
