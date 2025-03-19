package handler

import (
	"exapp-go/config"
	"exapp-go/pkg/utils"
	"testing"
)

func TestService_updateUserTokenBalance(t *testing.T) {
	utils.WorkInProjectPath("exapp-go")
	config.Load("config/config_dev.yaml")
	service, err := NewService()
	if err != nil {
		t.Fatalf("NewService failed: %v", err)
	}
	service.updateUserTokenBalance("abcabcabcabc", "active")
}
