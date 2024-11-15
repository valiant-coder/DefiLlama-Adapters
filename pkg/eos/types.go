package eos

import (
	"github.com/eoscanada/eos-go"
)

type AccountManager struct {
	eosClient *eos.API
	config    *Config
}

type Config struct {
	ChainID         string // chain id
	URL             string // EOS api url
	BaseAccountName string // base account name
	CreatorAccount  string // creator account
	CreatorPrivKey  string // creator private key
}

type AccountKeys struct {
	SocialPrivKey string // social private key
	HashPrivKey   string // hash private key
}
