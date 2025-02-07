package handler

import (
	"context"
	"encoding/json"
	"exapp-go/pkg/hyperion"
	"fmt"
	"log"
)

func (s *Service) handleNewAccount(action hyperion.Action) error {
	var data struct {
		ID      string `json:"id"`
		Account string `json:"account"`
		Pubkey  string `json:"pubkey"`
	}
	if err := json.Unmarshal(action.Act.Data, &data); err != nil {
		log.Printf("Unmarshal new account failed: %v", err)
		return nil
	}

	ctx := context.Background()
	credential, err := s.repo.GetUserCredentialByPubkey(ctx, data.Pubkey)
	if err != nil {
		log.Printf("Get user credential by pubkey failed: %v", err)
		return nil
	}

	credential.EOSAccount = data.Account
	credential.EOSPermissions = "active,owner"
	credential.BlockNumber = action.BlockNum
	err = s.repo.UpdateUserCredential(ctx, credential)
	if err != nil {
		log.Printf("Update user credential failed: %v-%v", data, err)
		return nil
	}

	return nil
}

func (s *Service) updateUserTokenBalance(account string) error {
	go s.publisher.PublishBalanceUpdate(account)

	return nil
}



func (s *Service) handleUpdateAuth(action hyperion.Action) error {
	ctx := context.Background()
	var data struct {
		Permission string `json:"permission"`
		Parent     string `json:"parent"`
		Auth       struct {
			Threshold int `json:"threshold"`
			Keys      []struct {
				Key    string `json:"key"`
				Weight int    `json:"weight"`
			} `json:"keys"`
		} `json:"auth"`
		Account string `json:"account"`
	}
	if err := json.Unmarshal(action.Act.Data, &data); err != nil {
		log.Printf("Unmarshal update auth failed: %v", err)
		return nil
	}

	credentials, err := s.repo.GetUserCredentialsByEOSAccount(ctx, data.Account)
	if err != nil {
		log.Printf("Get user credentials by eos account failed: %v", err)
		return nil
	}
	if len(credentials) == 0 {
		log.Printf("No user credentials found for account: %v", data.Account)
		return nil
	}

	keys := make(map[string]int)
	for _, key := range data.Auth.Keys {
		keys[key.Key] = key.Weight
	}

	for _, credential := range credentials {
		if _, ok := keys[credential.PublicKey]; ok {
			if credential.EOSPermissions == "" {
				credential.EOSPermissions = data.Permission
			} else {
				credential.EOSPermissions = fmt.Sprintf("%s,%s", credential.EOSPermissions, data.Permission)
			}
			credential.EOSAccount = data.Account
			credential.BlockNumber = action.BlockNum
			err = s.repo.UpdateUserCredential(ctx, credential)
			if err != nil {
				log.Printf("Update user credential failed: %v-%v", data, err)
				return nil
			}
		}else {
			log.Printf("No key found for credential: %v", credential.PublicKey)
			err = s.repo.DeleteUserCredential(ctx, credential)
			if err != nil {
				log.Printf("Delete user credential failed: %v-%v", data, err)
				return nil
			}
		}
	}

	return nil
}
