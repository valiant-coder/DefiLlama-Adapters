package onedex

import (
	"context"
	"testing"
)

func TestBridgeClient_GetDepositAddress(t *testing.T) {
	client := NewBridgeClient("",
		"",
		"",
		"",
	)
	address, err := client.GetDepositAddress(context.Background(), RequestDepositAddress{
		PermissionID: 0,
		Recipient:    "",
		Remark:       "",
	})
	if err != nil {
		t.Fatalf("Failed to get deposit address: %v", err)
	}
	t.Logf("Deposit address: %s", address)
}

func TestBridgeClient_MappingAddress(t *testing.T) {
	client := NewBridgeClient("",
		"",
		"",
		"",
	)
	resp, err := client.MappingAddress(context.Background(), MappingAddrRequest{
		PermissionID:         0,
		RecipientAddress:     "",
		Remark:               "",
		AssignDepositAddress: "",
	})
	if err != nil {
		t.Fatalf("Failed to mapping address: %v", err)
	}
	t.Logf("Mapping address: %s", resp.TransactionID)
}
