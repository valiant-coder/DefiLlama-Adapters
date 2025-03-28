package handler

import (
	"context"
	"encoding/json"
	"exapp-go/internal/db/db"
	"exapp-go/pkg/eos/onedex"
	"exapp-go/pkg/hyperion"
	"log"
	"strings"

	"github.com/shopspring/decimal"
	"github.com/spf13/cast"
)

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
		BlockNum:     action.BlockNum,
	}

	if err := s.repo.UpsertChain(ctx, chain); err != nil {
		log.Printf("failed to upsert chain: %v", err)
		return nil
	}
	return nil
}

/*
	"data": {
	    "token_id": "2",
	    "chain_ids": [
	        "2"
	    ],
	    "address": "18DA2F2617F5267C7BF35ECCB4F3D6FCD380F5A3",
	    "withdraw_helper": "69424744F1B0DAAFD4208B6A49CB03727A227E12",
	    "precision": 18,
	    "withdraw_method": 1
	}
*/
func (s *Service) handleMapXSAT(action hyperion.Action) error {
	ctx := context.Background()
	var data struct {
		TokenID        string   `json:"token_id"`
		ChainIDs       []string `json:"chain_ids"`
		Address        string   `json:"address"`
		WithdrawHelper string   `json:"withdraw_helper"`
		Precision      int32    `json:"precision"`
		WithdrawMethod int32    `json:"withdraw_method"`
	}
	if err := json.Unmarshal(action.Act.Data, &data); err != nil {
		log.Printf("failed to unmarshal map xsat data: %v", err)
		return nil
	}

	ondexTokens, err := onedex.GetOneDexSupportTokens(ctx, s.eosCfg.NodeURL, s.oneDexCfg.PortalContract)
	if err != nil {
		log.Printf("failed to get ondex tokens: %v", err)
		return nil
	}

	tokenID := cast.ToUint64(data.TokenID)

	var tokenSymbol string
	for _, token := range ondexTokens {
		if token.ID == tokenID {
			tokenSymbol = token.Symbol
			break
		}
	}
	if tokenSymbol == "" {
		log.Printf("failed to find token symbol: %v", tokenID)
		return nil
	}

	token, err := s.repo.GetToken(ctx, tokenSymbol)
	if err != nil {
		log.Printf("failed to get token: %v", err)
		return nil
	}

	chainIDs := make([]uint8, len(data.ChainIDs))
	for i, chainID := range data.ChainIDs {
		chainIDs[i] = cast.ToUint8(chainID)
	}
	chains, err := s.repo.GetChains(ctx, chainIDs)
	if err != nil {
		log.Printf("failed to get chains: %v", err)
		return nil
	}

	var chainInfos []db.ChainInfo
	for _, chain := range chains {
		permissionID := chain.PermissionID
		depositByBTCBridge := false
		if tokenSymbol == "BTC" && chain.ChainName != "exsat" {
			permissionID = 21000000
			depositByBTCBridge = true
		}
		exsatTokenAddress := data.Address
		if !strings.HasPrefix(exsatTokenAddress, "0x") {
			exsatTokenAddress = "0x" + exsatTokenAddress
		}

		chainInfos = append(chainInfos, db.ChainInfo{
			ChainID:            chain.ChainID,
			ChainName:          chain.ChainName,
			PermissionID:       permissionID,
			DepositByBTCBridge: depositByBTCBridge,

			WithdrawFee:       token.WithdrawFee,
			MinWithdrawAmount: token.WithdrawFee,
			MinDepositAmount:  token.WithdrawFee,

			ExsatWithdrawFee:      decimal.Zero,
			ExsatMinDepositAmount: token.WithdrawFee,
			ExsatTokenAddress:     exsatTokenAddress,
			ExsatTokenDecimals:    uint8(data.Precision),
		})
	}

	token.Chains = chainInfos
	token.BlockNum = action.BlockNum
	if err := s.repo.UpsertToken(ctx, token); err != nil {
		log.Printf("failed to upsert token: %v", err)
		return nil
	}

	return nil

}
