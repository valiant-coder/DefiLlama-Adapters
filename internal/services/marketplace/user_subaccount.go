package marketplace

import (
	"context"
	"crypto/rand"
	"exapp-go/internal/db/db"
	"exapp-go/internal/entity"
	"exapp-go/internal/errno"
	"fmt"
	"math/big"

	"log"

	"sync"

	"golang.org/x/sync/errgroup"
	"gorm.io/datatypes"
)

const (
	apiKeyLength = 36
	apiKeyPrefix = "k-"
)

// generateAPIKey generates a random API key with k- prefix followed by 36 random alphanumeric characters
func generateAPIKey() (string, error) {
	const chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	result := apiKeyPrefix

	for i := 0; i < apiKeyLength; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
		if err != nil {
			return "", err
		}
		result += string(chars[num.Int64()])
	}

	return result, nil
}

// AddSubAccount creates a new sub-account for the user
func (s *UserService) AddSubAccount(ctx context.Context, uid string, req entity.ReqAddSubAccount) (*entity.RespAddSubAccount, error) {
	user, err := s.repo.GetUser(ctx, uid)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if user.IsEVMUser() {
		return nil, errno.DefaultParamsError("EVM user cannot create sub-account")
	}
	// Generate API key
	apiKey, err := generateAPIKey()
	if err != nil {
		return nil, fmt.Errorf("failed to generate API key: %w", err)
	}

	// Create sub-account in database
	subAccount := &db.UserSubAccount{
		UID:        uid,
		Name:       req.Name,
		APIKey:     apiKey,
		PublicKeys: datatypes.NewJSONSlice([]string{}),
	}

	err = s.repo.CreateUserSubAccount(ctx, subAccount)
	if err != nil {
		return nil, fmt.Errorf("failed to create sub-account: %w", err)
	}

	return &entity.RespAddSubAccount{
		SID:    subAccount.SID,
		APIKey: apiKey,
	}, nil
}

// GetSubAccounts retrieves all sub-accounts for a user
func (s *UserService) GetSubAccounts(ctx context.Context, uid string) ([]*entity.SubAccountInfo, error) {
	subAccounts, err := s.repo.GetUserSubAccounts(ctx, uid)
	if err != nil {
		return nil, fmt.Errorf("failed to get sub-accounts: %w", err)
	}

	// Create a result slice with the same capacity as subAccounts
	result := make([]*entity.SubAccountInfo, len(subAccounts))

	// Use a wait group to wait for all goroutines to complete
	var wg sync.WaitGroup
	wg.Add(len(subAccounts))

	// Create a mutex to protect concurrent writes to the result slice
	var mu sync.Mutex

	// Use error group to handle errors from goroutines
	g, ctx := errgroup.WithContext(ctx)

	// Process each sub-account in parallel
	for i, sa := range subAccounts {
		i, sa := i, sa // Create local variables to avoid closure problems

		g.Go(func() error {
			defer wg.Done()

			// Fetch balance in the goroutine
			subAccountBalance, err := s.GetUserSubaccountBalances(ctx, sa.EOSAccount, sa.Permission)
			if err != nil {
				log.Printf("failed to get sub-account balance: %v", err)
				subAccountBalance = []entity.SubAccountBalance{}
			}

			// Create the SubAccountInfo object
			subAccountInfo := &entity.SubAccountInfo{
				SID:        sa.SID,
				Name:       sa.Name,
				EOSAccount: sa.EOSAccount,
				Permission: sa.Permission,
				APIKey:     sa.APIKey,
				PublicKeys: sa.PublicKeys,
				Balances:   subAccountBalance,
			}

			// Safely assign to the result slice at the correct index
			mu.Lock()
			result[i] = subAccountInfo
			mu.Unlock()

			return nil
		})
	}

	// Wait for all goroutines to complete
	wg.Wait()

	// Check if any error occurred in the goroutines
	if err := g.Wait(); err != nil {
		return nil, fmt.Errorf("error fetching sub-account balances: %w", err)
	}

	return result, nil
}

// DeleteSubAccount deletes a sub-account by name
func (s *UserService) DeleteSubAccount(ctx context.Context, req entity.ReqDeleteSubAccount) (*entity.RespDeleteSubAccount, error) {
	err := s.repo.DeleteUserSubAccount(ctx, req.SID)
	if err != nil {
		return nil, fmt.Errorf("failed to delete sub-account: %w", err)
	}

	return &entity.RespDeleteSubAccount{
		Success: true,
	}, nil
}
