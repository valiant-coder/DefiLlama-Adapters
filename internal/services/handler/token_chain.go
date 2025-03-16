package handler

import (
	"context"
	"encoding/json"
	"exapp-go/internal/db/db"
	"exapp-go/pkg/hyperion"
	"log"

	eosgo "github.com/eoscanada/eos-go"
	"github.com/spf13/cast"
)

func (s *Service) handleCreateToken(action hyperion.Action) error {
	ctx := context.Background()

	var data struct {
		Contract      string `json:"contract"`
		MaximumSupply string `json:"maximum_supply"`
		WithdrawFee   string `json:"withdraw_fee"`
	}
	if err := json.Unmarshal(action.Act.Data, &data); err != nil {
		log.Printf("failed to unmarshal create token data: %v", err)
		return nil
	}

	asset, err := eosgo.NewAssetFromString(data.MaximumSupply)
	if err != nil {
		log.Printf("failed to parse maximum supply: %v", err)
		return nil
	}

	token := &db.Token{
		EOSContractAddress: data.Contract,
		Symbol:             asset.Symbol.Symbol,
		Name:               asset.Symbol.Symbol,
		BlockNum:           action.BlockNum,
		Decimals:           asset.Symbol.Precision,
	}

	if err := s.repo.UpsertToken(ctx, token); err != nil {
		log.Printf("failed to upsert token: %v", err)
		return nil
	}
	return nil
}



func (s *Service) handleAddXSATChain(action hyperion.Action) error {
	ctx := context.Background()
	var data struct {
		ChainID      string `json:"chain_id"`
		PermissionID string `json:"permission_id"`
		ChainName    string `json:"chain_name"`
	}
	if err := json.Unmarshal(action.Act.Data, &data); err != nil {
		log.Printf("failed to unmarshal add xsat chain data: %v", err)
		return nil
	}

	chain := &db.Chain{
		ChainID:      cast.ToUint8(data.ChainID),
		PermissionID: cast.ToUint64(data.PermissionID),
		ChainName:    data.ChainName,
	}

	if err := s.repo.UpsertChain(ctx, chain); err != nil {
		log.Printf("failed to upsert chain: %v", err)
		return nil
	}
	return nil
}

