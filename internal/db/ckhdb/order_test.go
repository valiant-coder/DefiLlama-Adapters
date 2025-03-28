package ckhdb

import (
	"context"
	"exapp-go/config"
	"exapp-go/pkg/utils"
	"fmt"
	"testing"
)

func TestGetOrdersCoinTotal(t *testing.T) {

	utils.WorkInProjectPath("exapp-go")
	config.Load("config/config.yaml")
	ckhRepo := New()

	orders, err := ckhRepo.GetOrdersCoinTotal(context.Background(), "2024-10-01", "2025-10-31")
	if err != nil {
		t.Error(err)
	}

	for _, ororder := range orders {
		fmt.Println(&ororder)
	}
}
