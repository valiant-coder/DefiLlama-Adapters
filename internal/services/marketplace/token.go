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
	supportChains := make(map[string][]entity.Chain)
	for _, token := range tokens {
		supportChains[token.Symbol] = append(supportChains[token.Symbol], entity.Chain{
			ChainName:         token.ChainName,
			MinDepositAmount:  token.ExsatDepositLimit.String(),
			MinWithdrawAmount: token.ExsatWithdrawMax.String(),
		})
	}
	for symbol, chains := range supportChains {
		supportTokens = append(supportTokens, entity.Token{
			Symbol:       symbol,
			SupportChain: chains,
		})
	}

	return supportTokens, nil
}
