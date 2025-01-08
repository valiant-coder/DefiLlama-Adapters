package marketplace

import (
	"context"
	"exapp-go/config"
	"exapp-go/pkg/queryparams"
	"exapp-go/pkg/utils"
	"log"
	"testing"
)

func TestOrderService_GetOpenOrders(t *testing.T) {
	utils.WorkInProjectPath("exapp-go")
	config.Load("config/config_dev.yaml")
	orderService := NewOrderService()
	q := queryparams.NewCustomQueryParams()
	q.Add("trader", "5yziydihchpp")
	q.Add("pool_id", "0")
	q.Add("side", "1")
	orders, total, err := orderService.GetOpenOrders(context.Background(), q)
	if err != nil {
		t.Errorf("OrderService.GetOpenOrders() error = %v", err)
	}
	log.Printf("OrderService.GetOpenOrders() = %v", orders)
	log.Printf("OrderService.GetOpenOrders() = %v", total)
}

func TestOrderService_GetHistoryOrders(t *testing.T) {
	utils.WorkInProjectPath("exapp-go")
	config.Load("config/config_dev.yaml")
	orderService := NewOrderService()
	q := queryparams.NewCustomQueryParams()
	q.Add("trader", "playfullion5")
	orders, total, err := orderService.GetHistoryOrders(context.Background(), q)
	if err != nil {
		t.Errorf("OrderService.GetHistoryOrders() error = %v", err)
	}
	log.Printf("OrderService.GetHistoryOrders() = %v", orders)
	log.Printf("OrderService.GetHistoryOrders() = %v", total)
}

func TestOrderService_GetHistoryOrderDetail(t *testing.T) {
	utils.WorkInProjectPath("exapp-go")
	config.Load("config/config_dev.yaml")
	orderService := NewOrderService()
	order, err := orderService.GetHistoryOrderDetail(context.Background(), "0-1-1")
	if err != nil {
		t.Errorf("OrderService.GetHistoryOrderDetail() error = %v", err)
	}
	log.Printf("OrderService.GetHistoryOrderDetail() = %v", order)
}
