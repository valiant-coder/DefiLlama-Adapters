package handler

import (
	"encoding/json"
	"exapp-go/pkg/hyperion"
	"fmt"
)

func (s *Service) handleCancelOrder(action hyperion.Action) error {
	var data struct {
		OrderID string `json:"order_id"`
	}

	if err := json.Unmarshal(action.Act.Data, &data); err != nil {
		return fmt.Errorf("unmarshal cancel order data failed: %w", err)
	}

	return nil
}
