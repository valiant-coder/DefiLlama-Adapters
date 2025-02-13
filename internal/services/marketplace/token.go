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
		var supportChains []entity.Chain

		for _, chain := range token.Chains {
			if chain.MinDepositAmount.LessThan(chain.ExsatMinDepositAmount) {
				chain.MinDepositAmount = chain.ExsatMinDepositAmount
			}
			supportChains = append(supportChains, entity.Chain{
				ChainName: chain.ChainName,
				ChainID:   chain.ChainID,

				MinDepositAmount:  chain.MinDepositAmount.String(),
				MinWithdrawAmount: chain.MinWithdrawAmount.String(),

				WithdrawFee:      chain.WithdrawalFee.String(),
				ExsatWithdrawFee: chain.ExsatWithdrawFee.String(),
				ExsatTokenAddress: chain.ExsatTokenAddress,
			})
		}
		supportTokens = append(supportTokens, entity.Token{
			Symbol:       token.Symbol,
			SupportChain: supportChains,
			Name:         token.Name,
			Decimals:     token.Decimals,
			EOSContract:  token.EOSContractAddress,
		})
	}
	return supportTokens, nil
}
