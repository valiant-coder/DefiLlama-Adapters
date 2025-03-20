package cron

import (
	"context"
	"exapp-go/config"
	"exapp-go/internal/db/db"
	"exapp-go/pkg/telegram"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"
)

// formatDepositAlert formats a deposit record into an alert message
func formatDepositAlert(deposit *db.DepositRecord, timeoutMinutes time.Duration) string {
	return fmt.Sprintf("⚠️ Pending Deposit Alert\n"+
		"ID: %d\n"+
		"UID: %s\n"+
		"Symbol: %s\n"+
		"Amount: %s\n"+
		"Chain: %s\n"+
		"Address: %s\n"+
		"Pending since: %s\n"+
		"Timeout: %s\n"+
		"Transaction Hash: %s",
		deposit.ID,
		deposit.UID,
		deposit.Symbol,
		deposit.Amount.String(),
		deposit.ChainName,
		deposit.DepositAddress,
		deposit.Time.Format(time.RFC3339),
		timeoutMinutes.String(),
		deposit.TxHash)
}

// formatWithdrawAlert formats a withdrawal record into an alert message
func formatWithdrawAlert(withdrawal *db.WithdrawRecord, timeoutMinutes time.Duration) string {
	return fmt.Sprintf("⚠️ Pending Withdrawal Alert\n"+
		"ID: %d\n"+
		"UID: %s\n"+
		"Symbol: %s\n"+
		"Amount: %s\n"+
		"Chain: %s\n"+
		"Fee: %s\n"+
		"Recipient: %s\n"+
		"Pending since: %s\n"+
		"Timeout: %s\n"+
		"Transaction Hash: %s",
		withdrawal.ID,
		withdrawal.UID,
		withdrawal.Symbol,
		withdrawal.Amount.String(),
		withdrawal.ChainName,
		withdrawal.Fee.Add(withdrawal.BridgeFee).String(),
		withdrawal.Recipient,
		withdrawal.WithdrawAt.Format(time.RFC3339),
		timeoutMinutes.String(),
		withdrawal.TxHash)
}

// sendBatchAlerts sends multiple alert messages in a single batch
func sendBatchAlerts(messages []string) error {
	if len(messages) == 0 {
		return nil
	}

	// Join messages with double newlines for better readability
	batchMsg := strings.Join(messages, "\n\n")
	return telegram.SendMsg(batchMsg)
}

func (s *Service) MonitorPendingRecords() {
	ctx := context.Background()
	timeoutMinutes := config.Conf().Monitor.DepositWithdrawTimeout
	if timeoutMinutes <= 0 {
		timeoutMinutes = 10 * time.Minute // default timeout 10 minutes
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	var depositAlerts, withdrawalAlerts []string

	// Process deposits in a goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		pendingDeposits, err := s.repo.GetAllPendingDepositRecords(ctx)
		if err != nil {
			log.Printf("failed to get pending deposits: %v", err)
			return
		}

		threshold := time.Now().Add(-timeoutMinutes)
		for _, deposit := range pendingDeposits {
			if deposit.Time.Before(threshold) {
				mu.Lock()
				depositAlerts = append(depositAlerts, formatDepositAlert(deposit, timeoutMinutes))
				mu.Unlock()
			}
		}
	}()

	// Process withdrawals in a goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		pendingWithdrawals, err := s.repo.GetAllPendingWithdrawRecords(ctx)
		if err != nil {
			log.Printf("failed to get pending withdrawals: %v", err)
			return
		}

		threshold := time.Now().Add(-timeoutMinutes)
		for _, withdrawal := range pendingWithdrawals {
			if withdrawal.WithdrawAt.Before(threshold) {
				mu.Lock()
				withdrawalAlerts = append(withdrawalAlerts, formatWithdrawAlert(withdrawal, timeoutMinutes))
				mu.Unlock()
			}
		}
	}()

	// Wait for both goroutines to complete
	wg.Wait()

	// Send alerts in batches
	const batchSize = 10
	// Send deposit alerts
	for i := 0; i < len(depositAlerts); i += batchSize {
		end := i + batchSize
		if end > len(depositAlerts) {
			end = len(depositAlerts)
		}
		if err := sendBatchAlerts(depositAlerts[i:end]); err != nil {
			log.Printf("failed to send deposit alerts batch: %v", err)
		}
	}

	// Send withdrawal alerts
	for i := 0; i < len(withdrawalAlerts); i += batchSize {
		end := i + batchSize
		if end > len(withdrawalAlerts) {
			end = len(withdrawalAlerts)
		}
		if err := sendBatchAlerts(withdrawalAlerts[i:end]); err != nil {
			log.Printf("failed to send withdrawal alerts batch: %v", err)
		}
	}
}
