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

	EOS struct {
		NodeURL         string `yaml:"node_url"`
		PayerAccount    string `yaml:"payer_account"`
		PayerPrivateKey string `yaml:"payer_private_key"`
	} `yaml:"eos"`

	Hyperion struct {
		Endpoint       string `yaml:"endpoint"`
		StreamEndpoint string `yaml:"stream_endpoint"`
	} `yaml:"hyperion"`

	Nsq struct {
		Nsqds     []string      `yaml:"nsqds"`
		Lookupd   string        `yaml:"lookupd"`
		LookupTTl time.Duration `yaml:"lookup_ttl"`
	} `yaml:"nsq"`
}

type JWTConfig struct {
	SecretKey string `mapstructure:"secret_key"`
	Realm     string `mapstructure:"realm"`
	// hour
	Timeout int `mapstructure:"timeout"`
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
