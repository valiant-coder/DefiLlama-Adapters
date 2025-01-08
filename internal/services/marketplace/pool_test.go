package marketplace

import (
	"context"
	"exapp-go/config"
	"exapp-go/pkg/queryparams"
	"exapp-go/pkg/utils"
	"log"
	"testing"
)

func TestPoolService_GetPools(t *testing.T) {
	utils.WorkInProjectPath("exapp-go")
	config.Load("config/config_dev.yaml")
	poolService := NewPoolService()
	q := queryparams.NewCustomQueryParams()
	q.Add("quote_coin", "spottedtoken-USDC")
	pools, total, err := poolService.GetPools(context.Background(), q)
	if err != nil {
		t.Errorf("PoolService.GetPools() error = %v", err)
	}
	log.Printf("PoolService.GetPools() = %v", pools)
	log.Printf("PoolService.GetPools() = %v", total)
}
