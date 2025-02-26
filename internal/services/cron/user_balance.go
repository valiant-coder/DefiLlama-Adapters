package cron

import (
	"context"
	"exapp-go/internal/db/db"
	"exapp-go/internal/services/marketplace"
	"log"
	"time"
)

func getRoundedHour(t time.Time) time.Time {
	minutes := t.Minute()
	if minutes >= 30 {
		return time.Date(t.Year(), t.Month(), t.Day(), t.Hour()+1, 0, 0, 0, t.Location())
	}
	return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), 0, 0, 0, t.Location())
}

func (s *Service) recordUserBalances() {
	ctx := context.Background()
	eosAccounts, err := s.repo.GetAllEOSAccounts(ctx)
	if err != nil {
		log.Printf("Failed to get EOS account list: %v", err)
		return
	}
	userService := marketplace.NewUserService()

	now := getRoundedHour(time.Now())
	for _, eosAccount := range eosAccounts {
		usdtAmount, err := userService.CalculateUserUSDTBalance(ctx, eosAccount.EOSAccount)
		if err != nil {
			log.Printf("Failed to calculate USDT balance for user %s: %v", eosAccount.EOSAccount, err)
			continue
		}

		record := &db.UserBalanceRecord{
			Time:       now,
			Account:    eosAccount.EOSAccount,
			UID:        eosAccount.UID,
			USDTAmount: usdtAmount,
		}

		if err := s.repo.CreateUserBalanceRecord(ctx, record); err != nil {
			log.Printf("Failed to save balance record for user %s: %v", eosAccount.EOSAccount, err)
		}
	}
}
