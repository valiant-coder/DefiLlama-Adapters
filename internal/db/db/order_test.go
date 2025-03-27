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

func TestGetOrdersCoinTotal(t *testing.T) {
	utils.WorkInProjectPath("exapp-go")
	config.Load("config/config.yaml")

	r := New()
	startTime := "2023-07-01 00:00:00"
	endTime := "2025-07-01 23:59:59"

	resp, err := r.GetOrdersCoinTotal(context.Background(), startTime, endTime)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(resp)
}
