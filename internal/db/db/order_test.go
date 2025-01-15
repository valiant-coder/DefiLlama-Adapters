package db

import (
	"context"
	"exapp-go/config"
	"exapp-go/pkg/utils"
	"fmt"
	"testing"
)

func TestRepo_GetOpenOrderMaxBlockNumber(t *testing.T) {
	utils.WorkInProjectPath("exapp-go")
	config.Load("config/config_testnet2.yaml")
	r := New()
	blockNumber, err := r.GetOpenOrderMaxBlockNumber(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(blockNumber)
}
