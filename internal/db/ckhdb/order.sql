CREATE TABLE IF NOT EXISTS history_orders (
    app String,
    create_tx_id String,
    create_block_num UInt64,
    cancel_tx_id String,
    cancel_block_num UInt64,
    pool_id UInt64,
    pool_symbol String,
    pool_base_coin String,
    pool_quote_coin String,
    order_id UInt64,
    order_cid String,
    trader String,
    permission String,
    type UInt8,
    price Decimal(36,18),
    avg_price Decimal(36,18),
    is_bid Bool,
    original_quantity Decimal(36,18),
    executed_quantity Decimal(36,18),
    status UInt8,
    is_market Bool,
    created_at DateTime,
    canceled_at DateTime
) ENGINE = ReplacingMergeTree()
PRIMARY KEY (pool_id, order_id, is_bid)
ORDER BY (pool_id, order_id, is_bid)
SETTINGS index_granularity = 8192;

