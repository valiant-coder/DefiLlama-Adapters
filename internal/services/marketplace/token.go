package marketplace

import (
	"context"
	"exapp-go/internal/db/db"
	"exapp-go/internal/entity"
)

type TokenService struct {
	repo *db.Repo
}

func NewTokenService() *TokenService {
	return &TokenService{
		repo: db.New(),
	}
}

func (s *TokenService) GetSupportTokens(ctx context.Context) ([]entity.Token, error) {
	tokens, err := s.repo.ListTokens(ctx)
	if err != nil {
		return nil, err
	}
	var supportTokens []entity.Token

	for _, token := range tokens {
		supportTokens = append(supportTokens, entity.TokenFromDB(token))
	}
	return supportTokens, nil
}
