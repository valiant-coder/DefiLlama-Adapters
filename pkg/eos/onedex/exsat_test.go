package onedex

import (
	"context"
	"testing"

	"github.com/eoscanada/eos-go"
)

func TestGetEvmEosTokenMapping(t *testing.T) {
	client := eos.New("http://44.223.68.11:8888")
	mappings, err := GetEvmEosTokenMapping(context.Background(), client, "erc2o.xsat")
	if err != nil {
		t.Fatalf("Failed to get evm eos token mapping: %v", err)
	}
	t.Logf("Evm eos token mapping: %v", mappings)
}
