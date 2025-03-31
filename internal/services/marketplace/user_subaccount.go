package marketplace

import (
	"context"
	"crypto/rand"
	"exapp-go/internal/db/db"
	"exapp-go/internal/entity"
	"fmt"
	"math/big"

	"log"

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
		APIKey: apiKey,
	}, nil
}

// GetSubAccounts retrieves all sub-accounts for a user
func (s *UserService) GetSubAccounts(ctx context.Context, uid string) ([]*entity.SubAccountInfo, error) {
	subAccounts, err := s.repo.GetUserSubAccounts(ctx, uid)
	if err != nil {
		return nil, fmt.Errorf("failed to get sub-accounts: %w", err)
	}

	var result []*entity.SubAccountInfo
	for _, sa := range subAccounts {
		subAccountBalance, err := s.GetUserSubaccountBalances(ctx, sa.EOSAccount, sa.Permission)
		if err != nil {
			log.Printf("failed to get sub-account balance: %v", err)
			continue

		}
		result = append(result, &entity.SubAccountInfo{
			Name:       sa.Name,
			EOSAccount: sa.EOSAccount,
			Permission: sa.Permission,
			APIKey:     sa.APIKey,
			PublicKeys: sa.PublicKeys,
			Balances:   subAccountBalance,
		})
	}

	return result, nil
}

// DeleteSubAccount deletes a sub-account by name
func (s *UserService) DeleteSubAccount(ctx context.Context, uid string, req entity.ReqDeleteSubAccount) (*entity.RespDeleteSubAccount, error) {
	err := s.repo.DeleteUserSubAccount(ctx, uid, req.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to delete sub-account: %w", err)
	}

	return &entity.RespDeleteSubAccount{
		Success: true,
	}, nil
}
