package handler

import (
	"context"
	"exapp-go/internal/db/db"
	"exapp-go/pkg/hyperion"
	"fmt"
	"log"
)

func (s *Service) updateUserTokenBalance(account string) error {
	hyperionClient := hyperion.NewClient(s.hyperionCfg.Endpoint)
	ctx := context.Background()
	tokens, err := hyperionClient.GetTokens(ctx, account)
	if err != nil {
		log.Printf("Get tokens failed: %v", err)
		return nil
	}
	if err := s.repo.DeleteUserBalance(ctx, account); err != nil {
		log.Printf("Delete user balance failed: %v-%v", account, err)
		return nil
	}

	for _, token := range tokens {
		userBalance := &db.UserBalance{
			Account: account,
			Coin:    fmt.Sprintf("%s-%s", token.Contract, token.Symbol),
			Balance: token.Amount,
		}
		if err := s.repo.UpsertUserBalance(ctx, userBalance); err != nil {
			log.Printf("Upsert user balance failed: %v", err)
			continue
		}
	}
	return nil
}
