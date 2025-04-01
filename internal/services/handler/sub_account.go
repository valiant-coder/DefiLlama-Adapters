package handler

import (
	"context"
	"encoding/json"
	"exapp-go/pkg/hyperion"
	"log"
)

/*
	                "data": {
	                    "maker": "playfullion1",
	                    "subaccount": "aaaaaf.x"
						"sid":"012312"
	                }
*/
func (s *Service) handleRegSubAccount(action hyperion.Action) error {
	ctx := context.Background()
	var data struct {
		Maker      string `json:"maker"`
		SubAccount string `json:"subaccount"`
		SID        string `json:"sid"`
	}
	if err := json.Unmarshal(action.Act.Data, &data); err != nil {
		log.Printf("failed to unmarshal reg sub account data: %v", err)
		return nil
	}

	subAccount, err := s.repo.GetUserSubAccountBySID(ctx, data.SID)
	if err != nil {
		log.Printf("failed to get sub account: %v", err)
		return nil
	}

	subAccount.EOSAccount = data.SubAccount
	subAccount.Permission = "trader"
	subAccount.BlockNumber = action.BlockNum

	err = s.repo.UpdateUserSubAccount(ctx, subAccount)
	if err != nil {
		log.Printf("failed to update sub account: %v", err)
		return nil
	}

	return nil
}
