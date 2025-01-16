
CREATE TABLE IF NOT EXISTS pool_stats (
    pool_id UInt64,
    base_coin String,
    quote_coin String,
    symbol String,
    high Decimal(36,18),
    low Decimal(36,18),
    trades UInt64,
    last_price Decimal(36,18),
    volume Decimal(36,18),
    quote_volume Decimal(36,18),
    change Decimal(36,18),
    change_rate Float64,
    timestamp DateTime,
) ENGINE = ReplacingMergeTree(timestamp)
ORDER BY pool_id
SETTINGS index_granularity = 8192;