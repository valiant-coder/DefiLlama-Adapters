CREATE TABLE IF NOT EXISTS trades (
    tx_id String,
    pool_id UInt64,
    taker String,
    maker String,
    maker_order_id UInt64,
    maker_order_cid String,
    taker_order_id UInt64,
    taker_order_cid String,
    price Decimal(36,18),
    taker_is_bid Bool,
    base_quantity Decimal(36,18),
    quote_quantity Decimal(36,18),
    taker_fee Decimal(36,18),
    maker_fee Decimal(36,18),
    time DateTime,
    block_number UInt64,
	global_sequence UInt64,
	created_at DateTime
) ENGINE = ReplacingMergeTree(created_at)
PARTITION BY toYYYYMM(time)
PRIMARY KEY (global_sequence)
ORDER BY (global_sequence,pool_id, taker,maker, time)
SETTINGS index_granularity = 8192;