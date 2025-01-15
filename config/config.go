package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Mysql struct {
		Host     string `yaml:"host"`
		Port     string `yaml:"port"`
		Database string `yaml:"database"`
		User     string `yaml:"user"`
		Pass     string `yaml:"pass"`
		Loc      string `yaml:"loc"`
		Migrate  bool   `yaml:"migrate"`
	} `yaml:"mysql"`

	HTTPS struct {
		Enabled  bool   `yaml:"enabled"`
		CertFile string `yaml:"cert_file"`
		KeyFile  string `yaml:"key_file"`
	} `yaml:"https"`

	WS struct {
		HTTPS struct {
			Enabled  bool   `yaml:"enabled"`
			CertFile string `yaml:"cert_file"`
			KeyFile  string `yaml:"key_file"`
		} `yaml:"https"`
	} `yaml:"ws"`

	ClickHouse struct {
		Host     string `yaml:"host"`
		Port     string `yaml:"port"`
		Database string `yaml:"database"`
		User     string `yaml:"user"`
		Pass     string `yaml:"pass"`
	} `yaml:"clickhouse"`

	Trace struct {
		Exporter       string `yaml:"exporter"`
		JaegerEndpoint string `yaml:"jaeger_endpoint"`
	} `yaml:"trace"`

	Redis struct {
		IsCluster bool `yaml:"is_cluster"`
		Cluster   struct {
			Urls []string `yaml:"urls"`
			User string   `yaml:"user"`
			Pass string   `yaml:"pass"`
		} `yaml:"cluster"`
		Host string `yaml:"host"`
		Port int    `yaml:"port"`
		DB   int    `yaml:"db"`
		Pass string `yaml:"pass"`
	} `yaml:"redis"`

	JWT JWTConfig `mapstructure:"jwt"`

	Oauth2 struct {
		Google struct {
			ClientID string `yaml:"client_id"`
		} `yaml:"google"`
		Apple struct {
			ClientID string `yaml:"client_id"`
		} `yaml:"apple"`
	} `yaml:"oauth2"`

	Eos EosConfig `yaml:"eos"`

	Hyperion HyperionConfig `yaml:"hyperion"`

	Nsq NsqConfig `yaml:"nsq"`

	Cdex CdexConfig `yaml:"cdex"`
}

type EosConfig struct {
	NodeURL         string `yaml:"node_url"`
	PayerAccount    string `yaml:"payer_account"`
	PayerPrivateKey string `yaml:"payer_private_key"`
	Bridge          string `yaml:"bridge"`
	Actor           string `yaml:"actor"`
	ActorPrivateKey string `yaml:"actor_private_key"`
	VaultEVMAddress string `yaml:"vault_evm_address"`
}

type JWTConfig struct {
	SecretKey string `yaml:"secret_key"`
	Realm     string `yaml:"realm"`
	// hour
	Timeout int `yaml:"timeout"`
}

type HyperionConfig struct {
	StartBlock     uint64 `yaml:"start_block"`
	BatchSize      int    `yaml:"batch_size"`
	Endpoint       string `yaml:"endpoint"`
	StreamEndpoint string `yaml:"stream_endpoint"`
}

type NsqConfig struct {
	Nsqds     []string      `yaml:"nsqds"`
	Lookupd   string        `yaml:"lookupd"`
	LookupTTl time.Duration `yaml:"lookup_ttl"`
}

type CdexConfig struct {
	DexContract     string `yaml:"dex_contract"`
	PoolContract    string `yaml:"pool_contract"`
	EXAppContract   string `yaml:"exapp_contract"`
	HistoryContract string `yaml:"history_contract"`
	EventContract   string `yaml:"event_contract"`

	BridgeContract string `yaml:"bridge_contract"`

	DefaultPoolTakerFeeRate float64 `yaml:"default_pool_taker_fee_rate"`
	DefaultPoolMakerFeeRate float64 `yaml:"default_pool_maker_fee_rate"`
}

var (
	config = new(Config)
)

func Conf() *Config {
	return config
}

func Load(addr string) error {
	data, err := os.ReadFile(addr)
	if err != nil {
		fmt.Printf("load config file error: %+v\n", err)
		return err
	}

	err = yaml.Unmarshal(data, config)
	if err != nil {
		fmt.Printf("unmarshal config file error: %+v\n", err)
		return err
	}

	fmt.Printf("load config file success: %+v\n", config)
	return nil
}
