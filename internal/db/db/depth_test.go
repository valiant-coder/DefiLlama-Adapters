package db

import (
	"context"
	"exapp-go/config"
	"exapp-go/pkg/utils"
	"fmt"
	"testing"
)

func TestRepo_UpdateDepth(t *testing.T) {
	utils.WorkInProjectPath("exapp-go")
	config.Load("config/config_dev.yaml")
	r := New()
	err := r.UpdateDepth(context.Background(), []UpdateDepthParams{
		{PoolID: 1, IsBuy: false, Price: 100.0, Amount: -20, },
	})
	if err != nil {
		t.Fatal(err)
	}
	depth, err := r.GetDepth(context.Background(), 1)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(depth)

}
