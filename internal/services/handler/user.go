package handler

import (
	"context"
	"encoding/json"
	"errors"
	"exapp-go/internal/db/db"
	"exapp-go/pkg/eos/onedex"
	"exapp-go/pkg/hyperion"
	"fmt"
	"log"
	"strings"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

func (s *Service) handleNewAccount(action hyperion.Action) error {
	var data struct {
		ID      string `json:"id"`
		Account string `json:"account"`
	}
	if err := json.Unmarshal(action.Act.Data, &data); err != nil {
		log.Printf("Unmarshal new account failed: %v", err)
		return nil
	}

	ctx := context.Background()

	user, err := s.repo.GetUser(ctx, data.ID)
	if err != nil {
		log.Printf("Get user by uid failed: %v", err)
		return nil
	}

	user.EOSAccount = data.Account
	user.Permission = "active"
	user.BlockNumber = action.BlockNum
	err = s.repo.UpdateUser(ctx, user)
	if err != nil {
		log.Printf("Update user failed: %v-%v", data, err)
		return nil
	}

	signupClient := onedex.NewSignupClient(
		s.eosCfg.NodeURL,
		s.oneDexCfg.SignUpContract,
		s.oneDexCfg.Actor,
		s.oneDexCfg.ActorPrivateKey,
		s.oneDexCfg.ActorPermission,
	)
	pubkey, err := signupClient.GetPubkeyByUID(ctx, data.ID)
	if err != nil {
		log.Printf("Get pubkey by uid failed: %v", err)
		return nil
	}
	credential, err := s.repo.GetUserCredentialByPubkey(ctx, pubkey)
	if err != nil {
		log.Printf("Get user credential by uid failed: %v", err)
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
	go s.publisher.PublishBalanceUpdate(fmt.Sprintf("%s@%s", account, permission))

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

func (s *Service) handleEVMTraderMap(action hyperion.Action) error {
	ctx := context.Background()
	var data struct {
		Trader struct {
			Actor      string `json:"actor"`
			Permission string `json:"permission"`
		} `json:"trader"`
		Address string `json:"address"`
	}
	if err := json.Unmarshal(action.Act.Data, &data); err != nil {
		log.Printf("Unmarshal evm trader map failed: %v", err)
		return nil
	}
	evmAddress := strings.ToLower("0x" + data.Address)
	user, err := s.repo.GetUserByEVMAddress(ctx, evmAddress)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			user = &db.User{
				Username:    evmAddress,
				LoginMethod: db.LoginMethodEVM,
				OauthID:     evmAddress,
				EVMAddress:  evmAddress,
				EOSAccount:  data.Trader.Actor,
				Permission:  data.Trader.Permission,
				BlockNumber: action.BlockNum,
			}

		} else {
			log.Printf("Get user by evm address failed: %v", err)
			return nil
		}
	}
	user.EOSAccount = data.Trader.Actor
	user.Permission = data.Trader.Permission
	user.BlockNumber = action.BlockNum
	err = s.repo.UpsertUser(ctx, user)
	if err != nil {
		log.Printf("Upsert user failed: %v", err)
		return nil
	}
	return nil

}
