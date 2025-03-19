package handler

import (
	"context"
	"encoding/json"
	"exapp-go/pkg/hyperion"
	"fmt"
	"log"
	"strings"

	"gorm.io/datatypes"
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

func (s *Service) updateUserTokenBalance(account string, permission string) error {
	if permission != "active" {
		account = fmt.Sprintf("%s@%s", account, permission)
	}
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

	// handle sub account
	if data.Permission != "active" && data.Permission != "owner" {
		subAccount, err := s.repo.GetUserSubAccountByEOSAccountAndPermission(ctx, data.Account, data.Permission)
		if err != nil {
			return nil
		}
		keys := make([]string, 0)
		for _, key := range subAccount.PublicKeys {
			keys = append(keys, key)
		}
		subAccount.PublicKeys = datatypes.JSONSlice[string](keys)
		err = s.repo.UpdateUserSubAccount(ctx, subAccount)
		if err != nil {
			log.Printf("Update user sub account failed: %v", err)
			return nil
		}
		return nil
	}

	keys := make([]string, 0)
	for _, key := range data.Auth.Keys {
		keys = append(keys, key.Key)
	}
	keysCredentials, err := s.repo.GetUserCredentialsByKeys(ctx, keys)
	if err != nil {
		log.Printf("Get user credentials by keys failed: %v", err)
		return nil
	}

	for _, keyCredential := range keysCredentials {
		if keyCredential.EOSPermissions == "" {
			keyCredential.EOSPermissions = data.Permission
		} else if !strings.Contains(keyCredential.EOSPermissions, data.Permission) {
			keyCredential.EOSPermissions = fmt.Sprintf("%s,%s", keyCredential.EOSPermissions, data.Permission)
		}
		keyCredential.BlockNumber = action.BlockNum
		keyCredential.EOSAccount = data.Account
		err = s.repo.UpdateUserCredential(ctx, keyCredential)
		if err != nil {
			log.Printf("Update user credential failed: %v-%v", data, err)
			return nil
		}
	}

	userCredentials, err := s.repo.GetUserCredentialsByEOSAccount(ctx, data.Account)
	if err != nil {
		log.Printf("Get user credentials by eos account failed: %v", err)
		return nil
	}
	if len(userCredentials) == 0 {
		log.Printf("No user credentials found for account: %v", data.Account)
		return nil
	}

	keysMap := make(map[string]int)
	for _, key := range data.Auth.Keys {
		keysMap[key.Key] = key.Weight
	}

	for _, credential := range userCredentials {
		if _, ok := keysMap[credential.PublicKey]; ok {
			if credential.EOSPermissions == "" {
				credential.EOSPermissions = data.Permission
			} else if !strings.Contains(credential.EOSPermissions, data.Permission) {
				credential.EOSPermissions = fmt.Sprintf("%s,%s", credential.EOSPermissions, data.Permission)
			}
			credential.BlockNumber = action.BlockNum
			err = s.repo.UpdateUserCredential(ctx, credential)
			if err != nil {
				log.Printf("Update user credential failed: %v-%v", data, err)
				return nil
			}
		} else {
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
