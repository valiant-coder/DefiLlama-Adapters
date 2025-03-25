package marketplace

import (
	"context"
	"exapp-go/config"
	"exapp-go/pkg/utils"
	"log"
	"testing"
)

func TestUserService_GetUserBalance(t *testing.T) {
	utils.WorkInProjectPath("exapp-go")
	config.Load("config/config_dev.yaml")
	userService := NewUserService()
	userBalances, err := userService.FetchUserBalanceByUID(context.Background(), "abcabcabcabc")
	if err != nil {
		t.Errorf("GetUserBalance failed: %v", err)
	}
	log.Printf("GetUserBalance = %v", userBalances)
}
