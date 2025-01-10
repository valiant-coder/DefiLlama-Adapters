package eos

import (
	"context"
	"encoding/hex"
	"fmt"
	"log"
	"testing"

	"github.com/eoscanada/eos-go"
)

func TestFeeDelegation(t *testing.T) {

	// Set up EOS node API
	api := eos.New("https://jungle4.cryptolions.io") // Use actual EOS node address
	ctx := context.Background()

	// Transaction parameters
	fromAccount := ""             // Sender account
	toAccount := ""               // Receiver account
	quantity := "1.0000 EOS"      // Transfer amount
	memo := "test fee delegation" // Memo
	userPrivateKey := ""          // User private key
	payerAccount := ""            // Payer account
	payerPrivateKey := ""         // Payer private key

	// 1. Create user signed transaction
	singleSignedTx, err := CreateUserSignedTransaction(
		ctx,
		api,
		fromAccount,
		toAccount,
		quantity,
		memo,
		userPrivateKey,
		payerAccount,
	)
	if err != nil {
		log.Fatalf("Failed to create transaction: %v", err)
	}

	singleSignedTxBytes, err := eos.MarshalBinary(singleSignedTx)
	if err != nil {
		log.Fatalf("Failed to marshal transaction: %v", err)
	}
	// 2. Sign and broadcast transaction by payer account
	resp, err := SignAndBroadcastByPayer(
		ctx,
		api,
		hex.EncodeToString(singleSignedTxBytes),
		payerPrivateKey,
	)
	if err != nil {
		log.Fatalf("Failed to broadcast transaction: %v", err)
	}

	// 3. Print transaction result
	fmt.Printf("Transaction successful!\n")
	fmt.Printf("Transaction ID: %s\n", resp.TransactionID)
	fmt.Printf("Block Number: %d\n", resp.BlockNum)

}
