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
	
	Nsq NsqConfig `yaml:"nsq"`
	
	JWT JWTConfig `mapstructure:"jwt"`
	
	Oauth2 struct {
		Google struct {
			ClientID string `yaml:"client_id"`
		} `yaml:"google"`
		Apple struct {
			ClientID string `yaml:"client_id"`
		} `yaml:"apple"`
		Telegram struct {
			BotToken string `yaml:"bot_token"`
		} `yaml:"telegram"`
	} `yaml:"oauth2"`
	
	Eos EosConfig `yaml:"eos"`

	Evm EVMConfig `yaml:"evm"`

	Faucet struct {
		Enabled              bool    `yaml:"enabled"`
		EVMRpcUrl            string  `yaml:"evm_rpc_url"`
		USDTTokenAddress     string  `yaml:"usdt_token_address"`
		TokenOwner           string  `yaml:"token_owner"`
		TokenOwnerPrivateKey string  `yaml:"token_owner_private_key"`
		Amount               float64 `yaml:"amount"`
	} `yaml:"faucet"`
	
	TradingCompetition struct {
		BeginTime         time.Time `yaml:"begin_time"`
		EndTime           time.Time `yaml:"end_time"`
		DailyPoints       []int     `yaml:"daily_points"`
		AccumulatedPoints []int     `yaml:"accumulated_points"`
		FaucetPoint       int       `yaml:"faucet_points"`
		Blacklist         []string  `yaml:"blacklist"`
	} `yaml:"trading_competition"`
	
	TelegramBot struct {
		Token  string `yaml:"token"`
		ChatID string `yaml:"chat_id"`
	} `yaml:"telegram_bot"`
	
	Monitor struct {
		DepositWithdrawTimeout time.Duration `yaml:"deposit_withdraw_timeout"`
	} `yaml:"monitor"`
	
	Invitation struct {
		Host string `yaml:"host"`
	} `yaml:"invitation"`
}

type EVMConfig struct {
	Ethscan struct {
		Endpoint string `yaml:"endpoint"`
		ApiKey   string `yaml:"api_key"`
	} `yaml:"ethscan"`

	ExsatNetwork struct {
		CurrencySymbol   string `yaml:"currency_symbol"`
		NetworkUrl       string `yaml:"network_url"`
		ChainId          int    `yaml:"chain_id"`
		NetworkName      string `yaml:"network_name"`
		BlockExplorerUrl string `yaml:"block_explorer_url"`
	} `yaml:"exsat_network"`
}



type NsqConfig struct {
	Nsqds     []string      `yaml:"nsqds"`
	Lookupd   string        `yaml:"lookupd"`
	LookupTTl time.Duration `yaml:"lookup_ttl"`
}

type JWTConfig struct {
	SecretKey string `yaml:"secret_key"`
	Realm     string `yaml:"realm"`
	// hour
	Timeout int `yaml:"timeout"`
}

type EosConfig struct {
	NodeURL         string         `yaml:"node_url"`
	PayerAccount    string         `yaml:"payer_account"`
	PayerPrivateKey string         `yaml:"payer_private_key"`
	Hyperion        HyperionConfig `yaml:"hyperion"`
	CdexConfig      CdexConfig     `yaml:"cdex"`
	OneDex          OneDexConfig   `yaml:"onedex"`
	Exsat           ExsatConfig    `yaml:"exsat"`
	Events          EventConfig    `yaml:"events"`
	
	PowerUp PowerUpConfig `yaml:"powerup"`
}

type PowerUpConfig struct {
	Enabled           bool    `yaml:"enabled"`
	NetEOS            uint64  `yaml:"net_eos"`
	CPUEOS            uint64  `yaml:"cpu_eos"`
	MaxPayment        uint64  `yaml:"max_payment"`
	CPUMonitorEnabled bool    `yaml:"cpu_monitor_enabled"`
	CPUThreshold      float64 `yaml:"cpu_threshold"` // Threshold percentage (0-100) to trigger powerup
}

type HyperionConfig struct {
	StartBlock       uint64 `yaml:"start_block"`
	BatchSize        int    `yaml:"batch_size"`
	SyncTradeHistory bool   `yaml:"sync_trade_history"`
	Endpoint         string `yaml:"endpoint"`
	StreamEndpoint   string `yaml:"stream_endpoint"`
}

type OneDexConfig struct {
	VaultEVMAddress          string  `yaml:"vault_evm_address"`
	BridgeContract           string  `yaml:"bridge_contract"`
	BridgeContractEVMAddress string  `yaml:"bridge_contract_evm_address"`
	Actor                    string  `yaml:"actor"`
	ActorPrivateKey          string  `yaml:"actor_private_key"`
	AppTakerFeeRate          float64 `yaml:"app_taker_fee_rate"`
	AppMakerFeeRate          float64 `yaml:"app_maker_fee_rate"`
	TokenContract            string  `yaml:"token_contract"`
	EVMAgentContract         string  `yaml:"evm_agent_contract"`  
}

type ExsatConfig struct {
	BridgeContract    string `yaml:"bridge_contract"`
	BTCBridgeContract string `yaml:"btc_bridge_contract"`
	BTCChainName      string `yaml:"btc_chain_name"`
}

type CdexConfig struct {
	DexContract     string `yaml:"dex_contract"`
	PoolContract    string `yaml:"pool_contract"`
	OneDexContract  string `yaml:"onedex_contract"`
	HistoryContract string `yaml:"history_contract"`
	EventContract   string `yaml:"event_contract"`
	
	DefaultPoolTakerFeeRate float64 `yaml:"default_pool_taker_fee_rate"`
	DefaultPoolMakerFeeRate float64 `yaml:"default_pool_maker_fee_rate"`
}

var (
	config = new(Config)
)

func Conf() *Config {
	if config.Eos.Events.LogNewAcc == "" {
		config.Eos.Events = DefaultEventConfig()
	}
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

type EventConfig struct {
	// Order related events
	EmitPlaced   string `yaml:"emit_placed"`
	EmitCanceled string `yaml:"emit_canceled"`
	EmitFilled   string `yaml:"emit_filled"`
	
	// Pool related events
	Create string `yaml:"create"`
	
	// Deposit and withdrawal related events
	// bridge deposit log
	DepositLog string `yaml:"deposit_log"`
	// exapp deposit log
	LogDeposit string `yaml:"log_deposit"`
	
	LogNewAcc   string `yaml:"log_new_acc"`
	LogWithdraw string `yaml:"log_withdraw"`
	WithdrawLog string `yaml:"withdraw_log"`
	// exapp eos native withdraw log
	LogSend string `yaml:"log_send"`
	
	SetMinAmt string `yaml:"set_min_amt"`
	
	SetPoolFeeRate string `yaml:"set_pool_fee_rate"`
	
	// token-chain events
	CreateToken  string `yaml:"create_token"`
	AddXSATChain string `yaml:"add_xsat_chain"`
	MapXSAT      string `yaml:"map_xsat"`

	LogNewTrader string `yaml:"log_new_trader"`
}

func DefaultEventConfig() EventConfig {
	return EventConfig{
		EmitPlaced:     "emitplaced",
		EmitCanceled:   "emitcanceled",
		EmitFilled:     "emitfilled",
		Create:         "create",
		DepositLog:     "depositlog",
		LogDeposit:     "logdeposit",
		LogNewAcc:      "lognewacc",
		LogWithdraw:    "logwithdraw",
		WithdrawLog:    "withdrawlog",
		LogSend:        "logsend",
		SetMinAmt:      "setminamt",
		SetPoolFeeRate: "setpfeerate",
		CreateToken:    "createtoken",
		AddXSATChain:   "addxsatchain",
		MapXSAT:        "mapxsat",
		LogNewTrader:   "lognewtrader",
	}
}
