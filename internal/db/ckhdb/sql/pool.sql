


CREATE TABLE IF NOT EXISTS pools (
    pool_id             UInt64,
    base_symbol         String,
    base_contract       String,
    base_coin           String,
    base_coin_precision UInt8,
    quote_symbol        String,
    quote_contract      String,
    quote_coin          String,
    quote_coin_precision UInt8,
    symbol              String,
    asking_time         DateTime,
    trading_time        DateTime,
    max_flct            UInt64,
    price_precision     UInt8,
    taker_fee_rate      UInt64,
    maker_fee_rate      UInt64,
    status              UInt8,
    created_at DateTime DEFAULT now()
) ENGINE = ReplacingMergeTree(created_at)
ORDER BY pool_id
PRIMARY KEY (pool_id);


ALTER TABLE pools ADD INDEX idx_base_coin base_coin TYPE minmax;
ALTER TABLE pools ADD INDEX idx_quote_coin quote_coin TYPE minmax;


CREATE TABLE IF NOT EXISTS pool_stats (
    pool_id UInt64,
    high Decimal(36,18),
    low Decimal(36,18),
    trades UInt64,
    volume Decimal(36,18),
    quote_volume Decimal(36,18),
    price_change Float64,
    timestamp DateTime,
) ENGINE = ReplacingMergeTree(timestamp)
ORDER BY pool_id
SETTINGS index_granularity = 8192;