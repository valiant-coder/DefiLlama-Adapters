package hyperion

import (
	"log"
	"testing"
)

func TestNewStreamClient(t *testing.T) {
	endpoint := "https://eos.hyperion.eosrio.io"
	client, err := NewStreamClient(endpoint)
	if err != nil { 
		t.Fatalf("NewStreamClient() error = %v", err)
	}

	actionCh, err := client.SubscribeAction(ActionStreamRequest{
		Account:  "",
		Contract: "",
		Action:   "",
		Filters:  []RequestFilter{},
		StartFrom: 0,
		ReadUntil: 0,
	})
	if err != nil {
		t.Fatalf("SubscribeAction() error = %v", err)
	}

	for action := range actionCh {
		log.Printf("action: %+v", action)
	}

	select {}
}
