package repair

import (
	"context"
	"exapp-go/config"
	"exapp-go/pkg/utils"
	"testing"
)

func TestServer_RepairPool(t *testing.T) {
	utils.WorkInProjectPath("exapp-go")
	config.Load("config/config_dev.yaml")
	server := NewRepairServer()
	err := server.RepairPool(context.Background(), 2)
	if err != nil {
		t.Fatal(err)
	}
}
