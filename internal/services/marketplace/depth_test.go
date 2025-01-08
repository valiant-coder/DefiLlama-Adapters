package marketplace

import (
	"context"
	"exapp-go/config"
	"exapp-go/pkg/utils"
	"testing"
)

func TestDepthService_GetDepth(t *testing.T) {
	utils.WorkInProjectPath("exapp-go")
	config.Load("config/config_dev.yaml")
	depthService := NewDepthService()

	depth, err := depthService.GetDepth(context.Background(), 0)
	if err != nil {
		t.Errorf("DepthService.GetDepth() error = %v", err)
	}
	t.Logf("DepthService.GetDepth() = %v", depth)
}
