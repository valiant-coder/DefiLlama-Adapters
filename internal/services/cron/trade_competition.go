package cron

import (
	"context"
	"exapp-go/config"
	"exapp-go/internal/db/db"
	"exapp-go/internal/services/marketplace"
	"fmt"
	"log"
	"sort"
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

func (s *Service) HandleTradeCompetition() {
	log.Printf("Calculate user balances")
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

	log.Printf("Calculate trade competition points")
	err = s.calculateTradeCompetitionPoints()
	if err != nil {
		log.Printf("Failed to calculate trade competition points: %v", err)
	}
}

func (s *Service) recordUserBalances() error {
	ctx := context.Background()
	eosAccounts, err := s.repo.GetAllEOSAccounts(ctx)
	if err != nil {
		log.Printf("Failed to get EOS account list: %v", err)
		return err
	}

	// Create user service instance
	userService := marketplace.NewUserService()
	now := getRoundedHour(time.Now())

	// Use channel to collect results
	type result struct {
		record *db.UserBalanceRecord
		err    error
	}
	resultChan := make(chan result, len(eosAccounts))

	// Use worker pool to control concurrency
	workerCount := 10
	if len(eosAccounts) < workerCount {
		workerCount = len(eosAccounts)
	}

	// Create task channel
	taskChan := make(chan db.EOSAccountInfo, len(eosAccounts))
	for _, account := range eosAccounts {
		taskChan <- account
	}
	close(taskChan)

	// Start worker goroutines
	for i := 0; i < workerCount; i++ {
		go func() {
			for account := range taskChan {
				usdtAmount, err := userService.CalculateUserUSDTBalance(ctx, account.EOSAccount)
				if err != nil {
					resultChan <- result{err: fmt.Errorf("failed to calculate USDT balance for user %s: %w", account.EOSAccount, err)}
					continue
				}

				record := &db.UserBalanceRecord{
					Time:       now,
					Account:    account.EOSAccount,
					UID:        account.UID,
					USDTAmount: usdtAmount,
				}
				resultChan <- result{record: record}
			}
		}()
	}

	// Collect results
	var records []*db.UserBalanceRecord
	var errors []error
	for i := 0; i < len(eosAccounts); i++ {
		res := <-resultChan
		if res.err != nil {
			errors = append(errors, res.err)
			continue
		}
		records = append(records, res.record)
	}

	// Batch save records
	if len(records) > 0 {
		if err := s.repo.BatchCreateUserBalanceRecords(ctx, records); err != nil {
			log.Printf("Failed to batch save balance records: %v", err)
			return err
		}
	}

	// Handle error logs together
	if len(errors) > 0 {
		log.Printf("Encountered %d errors while processing user balances:", len(errors))
		for _, err := range errors {
			log.Printf("Error: %v", err)
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

	type UserProfitData struct {
		account    string
		uid        string
		lastAmount decimal.Decimal
		profit     decimal.Decimal
	}
	userProfits := make(map[string]*UserProfitData, len(records))

	for _, record := range records {
		userData, exists := userProfits[record.UID]
		if !exists {
			userProfits[record.UID] = &UserProfitData{
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

	profitRecords := make([]*db.UserDayProfitRecord, 0, len(userProfits))
	for _, userData := range userProfits {
		if userData.profit.Equal(decimal.Zero) {
			continue
		}
		profitRecords = append(profitRecords, &db.UserDayProfitRecord{
			Time:    dayStart,
			Account: userData.account,
			UID:     userData.uid,
			Profit:  userData.profit,
		})
	}

	if len(profitRecords) > 0 {
		if err := s.repo.BatchUpsertUserDayProfitRecords(ctx, profitRecords); err != nil {
			log.Printf("Failed to batch upsert day profit records: %v", err)
			return err
		}
	}
	return nil
}

func (s *Service) calculateUserAccumulatedProfit() error {
	ctx := context.Background()
	beginTime := config.Conf().TradingCompetition.BeginTime
	endTime := config.Conf().TradingCompetition.EndTime

	now := time.Now()
	if now.Before(beginTime) || now.After(endTime.Add(30*time.Second)) {
		log.Printf("Current time %v is not in accumulated profit time range [%v, %v]", now, beginTime, endTime)
		return nil
	}

	eosAccounts, err := s.repo.GetAllEOSAccounts(ctx)
	if err != nil {
		return fmt.Errorf("failed to get EOS accounts: %w", err)
	}

	uidToAccount := make(map[string]db.EOSAccountInfo, len(eosAccounts))
	uids := make([]string, len(eosAccounts))
	for i, account := range eosAccounts {
		uidToAccount[account.UID] = account
		uids[i] = account.UID
	}

	records, err := s.repo.GetUserBalanceRecordsInTimeRangeForUIDs(ctx, uids, beginTime, endTime)
	if err != nil {
		return fmt.Errorf("failed to get balance records: %w", err)
	}

	userRecords := make(map[string][]db.UserBalanceRecord)
	for _, record := range records {
		userRecords[record.UID] = append(userRecords[record.UID], record)
	}

	var profitRecords []*db.UserAccumulatedProfitRecord
	for uid, records := range userRecords {
		if len(records) < 2 {
			continue
		}

		sort.Slice(records, func(i, j int) bool {
			return records[i].Time.Before(records[j].Time)
		})

		profit := records[len(records)-1].USDTAmount.Sub(records[0].USDTAmount)
		account := uidToAccount[uid]

		profitRecords = append(profitRecords, &db.UserAccumulatedProfitRecord{
			BeginTime: beginTime,
			EndTime:   endTime,
			Account:   account.EOSAccount,
			UID:       uid,
			Profit:    profit,
		})
	}

	if len(profitRecords) > 0 {
		if err := s.repo.BatchUpsertUserAccumulatedProfitRecords(ctx, profitRecords); err != nil {
			log.Printf("Failed to batch upsert accumulated profit records: %v", err)
			return err
		}
	}

	return nil
}

func (s *Service) calculateTradeCompetitionPoints() error {
	ctx := context.Background()
	now := time.Now().UTC()
	yesterday := now.Add(-24 * time.Hour)
	yesterdayStart := time.Date(yesterday.Year(), yesterday.Month(), yesterday.Day(), 0, 0, 0, 0, time.UTC)
	yesterdayEnd := yesterdayStart.Add(24 * time.Hour)

	competitionBeginTime := config.Conf().TradingCompetition.BeginTime
	competitionEndTime := config.Conf().TradingCompetition.EndTime

	if now.Before(competitionBeginTime) || now.After(competitionEndTime.Add(30*time.Second)) {
		log.Printf("Current time %v is not in competition time range [%v, %v]", now, competitionBeginTime, competitionEndTime)
		return nil
	}

	dayProfitRecords, err := s.repo.GetUserDayProfitRanking(ctx, yesterdayStart, len(config.Conf().TradingCompetition.DailyPoints))
	if err != nil {
		return fmt.Errorf("failed to get day profit ranking: %w", err)
	}

	var records []*db.TradeCompetitionRecord
	for i, record := range dayProfitRecords {
		if i >= len(config.Conf().TradingCompetition.DailyPoints) {
			break
		}
		points := config.Conf().TradingCompetition.DailyPoints[i]
		records = append(records, &db.TradeCompetitionRecord{
			UID:       record.UID,
			Points:    points,
			BeginTime: yesterdayStart,
			EndTime:   yesterdayEnd,
		})
	}

	if now.After(competitionEndTime){
		accumulatedRecords, err := s.repo.GetUserAccumulatedProfitRanking(ctx, competitionBeginTime, competitionEndTime, len(config.Conf().TradingCompetition.AccumulatedPoints))
		if err != nil {
			return fmt.Errorf("failed to get accumulated profit ranking: %w", err)
		}

		for i, record := range accumulatedRecords {
			if i >= len(config.Conf().TradingCompetition.AccumulatedPoints) {
				break
			}
			points := config.Conf().TradingCompetition.AccumulatedPoints[i]
			records = append(records, &db.TradeCompetitionRecord{
				UID:       record.UID,
				Points:    points,
				BeginTime: competitionBeginTime,
				EndTime:   competitionEndTime,
			})
		}
	}

	if len(records) > 0 {
		for _, record := range records {
			err := s.repo.UpsertTradeCompetitionRecord(ctx, record)
			if err != nil {
				log.Printf("Failed to upsert trade competition record for uid %s: %v", record.UID, err)
			}
		}
	}

	return nil
}
