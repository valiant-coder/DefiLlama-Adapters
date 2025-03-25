package onedex

import (
	"context"
	"fmt"
	"testing"
)

func TestBTCBridgeClient_MappingAddress(t *testing.T) {
	client := NewBTCBridgeClient(
		"http://44.223.68.11:8888",
		"brdgmng.xsat",
		"asdfasdfasdf",
		"",
	)
	resp, err := client.MappingAddress(context.Background(), BTCMappingAddrRequest{
		RecipientAddress:     "asdfasdfasdf",
		Remark:               "test3test3",
		AssignDepositAddress: "",
	})
	if err != nil {
		t.Fatalf("Failed to mapping address: %v", err)
	}
	fmt.Println(resp.TransactionID)
	t.Logf("Mapping address response: %+v", resp)
}


func TestBTCBridgeClient_GetDepositAddress(t *testing.T) {
	client := NewBTCBridgeClient(
		"http://44.223.68.11:8888",
		"brdgmng.xsat",
		"",
		"",
	)
	resp, err := client.GetDepositAddress(context.Background(), RequestBTCDepositAddress{
		Remark:              "test3test3",
		RecipientEVMAddress: "",
	})
	if err != nil {
		t.Fatalf("Failed to get deposit address: %v", err)
	}
	fmt.Println(resp)
}
