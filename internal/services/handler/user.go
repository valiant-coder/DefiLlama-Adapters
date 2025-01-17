package handler

import (
	"context"
	"encoding/json"
	"exapp-go/pkg/hyperion"
	"log"
)

func (s *Service) handleNewAccount(action hyperion.Action) error {
	var data struct {
		ID      string `json:"id"`
		Account string `json:"account"`
		Pubkey  string `json:"pubkey"`
	}
	if err := json.Unmarshal(action.Act.Data, &data); err != nil {
		log.Printf("Unmarshal new account failed: %v", err)
		return nil
	}

	ctx := context.Background()
	credential, err := s.repo.GetUserCredentialByPubkey(ctx, data.Pubkey)
	if err != nil {
		log.Printf("Get user credential by pubkey failed: %v", err)
		return nil
	}

	credential.EOSAccount = data.Account
	credential.EOSPermissions = "active,owner"
	err = s.repo.UpdateUserCredential(ctx, credential)
	if err != nil {
		log.Printf("Update user credential failed: %v-%v", data, err)
		return nil
	}

	return nil
}

func (s *Service) updateUserTokenBalance(account string) error {
	go s.publisher.PublishBalanceUpdate(account)

	return nil
}
