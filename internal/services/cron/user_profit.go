package cron

import (
	"context"
	"exapp-go/config"
	"exapp-go/internal/db/db"
	"exapp-go/internal/services/marketplace"
	"fmt"
	"log"
	"time"

	"github.com/shopspring/decimal"
)

func getRoundedHour(t time.Time) time.Time {
	minutes := t.Minute()
	if minutes >= 30 {
		return time.Date(t.Year(), t.Month(), t.Day(), t.Hour()+1, 0, 0, 0, t.Location())
	}
	return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), 0, 0, 0, t.Location())
}

func (s *Service) HandleUserProfit() {
	err := s.recordUserBalances()
	if err != nil {
		log.Printf("Failed to record user balances: %v", err)
		return
	}

	log.Printf("Calculate user day profit")
	err = s.calculateUserDayProfit()
	if err != nil {
		log.Printf("Failed to calculate user day profit: %v", err)
	}

	log.Printf("Calculate user accumulated profit")
	err = s.calculateUserAccumulatedProfit()
	if err != nil {
		log.Printf("Failed to calculate user accumulated profit: %v", err)
	}
}

func (s *Service) recordUserBalances() error {
	ctx := context.Background()
	eosAccounts, err := s.repo.GetAllEOSAccounts(ctx)
	if err != nil {
		log.Printf("Failed to get EOS account list: %v", err)
		return err
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

	return nil
}

func (s *Service) calculateUserDayProfit() error {
	ctx := context.Background()
	now := time.Now().UTC()
	dayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	records, err := s.repo.GetUserBalanceRecordsByTimeRange(ctx, dayStart, now)
	if err != nil {
		log.Printf("Failed to get user balance records: %v", err)
		return err
	}

	userProfits := make(map[string]*struct {
		account    string
		uid        string
		lastAmount decimal.Decimal
		profit     decimal.Decimal
	})

	for _, record := range records {
		userData, exists := userProfits[record.UID]
		if !exists {
			userProfits[record.UID] = &struct {
				account    string
				uid        string
				lastAmount decimal.Decimal
				profit     decimal.Decimal
			}{
				account:    record.Account,
				uid:        record.UID,
				lastAmount: record.USDTAmount,
				profit:     decimal.Zero,
			}
			continue
		}

		profit := record.USDTAmount.Sub(userData.lastAmount)
		userData.profit = userData.profit.Add(profit)
		userData.lastAmount = record.USDTAmount
	}

	for _, userData := range userProfits {
		profitRecord := &db.UserDayProfitRecord{
			Time:    dayStart,
			Account: userData.account,
			UID:     userData.uid,
			Profit:  userData.profit,
		}
		if err := s.repo.UpsertUserDayProfitRecord(ctx, profitRecord); err != nil {
			log.Printf("Failed to upsert user day profit record: %v", err)
			continue
		}
	}

	return nil
}

func (s *Service) calculateUserAccumulatedProfit() error {
	ctx := context.Background()
	beginTime := config.Conf().AccumulatedProfit.BeginTime
	endTime := config.Conf().AccumulatedProfit.EndTime

	now := time.Now()
	if now.Before(beginTime) || now.After(endTime) {
		log.Printf("Current time %v is not in accumulated profit time range [%v, %v]", now, beginTime, endTime)
		return nil
	}

	eosAccounts, err := s.repo.GetAllEOSAccounts(ctx)
	if err != nil {
		return fmt.Errorf("failed to get EOS accounts: %w", err)
	}

	for _, eosAccount := range eosAccounts {
		records, err := s.repo.GetUserBalanceRecordsInTimeRange(ctx, eosAccount.UID, beginTime, endTime)
		if err != nil {
			log.Printf("Failed to get balance records for user %s: %v", eosAccount.EOSAccount, err)
			continue
		}

		if len(records) == 0 {
			continue
		}

		var profit decimal.Decimal
		if len(records) >= 2 {
			profit = records[len(records)-1].USDTAmount.Sub(records[0].USDTAmount)
		}

		record := &db.UserAccumulatedProfitRecord{
			BeginTime: beginTime,
			EndTime:   endTime,
			Account:   eosAccount.EOSAccount,
			UID:       eosAccount.UID,
			Profit:    profit,
		}

		if err := s.repo.UpsertUserAccumulatedProfitRecord(ctx, record); err != nil {
			log.Printf("Failed to upsert accumulated profit record for user %s: %v", eosAccount.EOSAccount, err)
			continue
		}
	}

	return nil
}
