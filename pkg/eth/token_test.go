package eth

import (
	"log"
	"testing"
)

func TestToken_MintERC20(t *testing.T) {
	client, err := NewClient("https://ethereum-sepolia-rpc.publicnode.com")
	if err != nil {
		log.Default().Printf("Failed to create client: %v", err)
		t.Fatalf("Failed to create client: %v", err)
	}

	token, err := NewToken("")
	if err != nil {
		log.Default().Printf("Failed to create token: %v", err)
		t.Fatalf("Failed to create token: %v", err)
	}

	txId, err := token.MintERC20(client,
		"0xB89B2B925fd2BA154051a5B77161EfB9AF1Fd7Fd",
		"",
		100,
		"",
		170,
	)
	if err != nil {
		log.Default().Printf("Failed to mint token: %v", err)
		t.Fatalf("Failed to mint token: %v", err)
	}

	log.Default().Printf("Minted token with txId: %s", txId)

}
