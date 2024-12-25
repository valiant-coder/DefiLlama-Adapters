package handler

import (
	"encoding/json"
	"exapp-go/pkg/hyperion"
	"fmt"
)

func (s *Service) handleMatchOrder(action hyperion.Action) error {
	var data struct {
		MakerOrderID string `json:"maker_order_id"`
		TakerOrderID string `json:"taker_order_id"`
		Price        string `json:"price"`
		Amount       string `json:"amount"`
	}

	if err := json.Unmarshal(action.Act.Data, &data); err != nil {
		return fmt.Errorf("unmarshal match order data failed: %w", err)
	}

	return nil
}
