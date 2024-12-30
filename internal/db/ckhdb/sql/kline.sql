CREATE TABLE IF NOT EXISTS klines (
    pool_id UInt64,
    interval_start DateTime,
    interval String,
    open_price AggregateFunction(argMin, Decimal(36,18), DateTime),    -- open
    max_price AggregateFunction(max, Decimal(36,18)),                  -- high
    min_price AggregateFunction(min, Decimal(36,18)),                  -- low
    close_price AggregateFunction(argMax, Decimal(36,18), DateTime),   -- close
    volume AggregateFunction(sum, Decimal(36,18)),                     -- volume
    quote_volume AggregateFunction(sum, Decimal(36,18)),              -- quote_volume
    trade_count AggregateFunction(count, UInt64)                      -- trade_count
) ENGINE = AggregatingMergeTree()
PARTITION BY toYYYYMM(interval_start)
ORDER BY (pool_id, interval, interval_start);


CREATE VIEW IF NOT EXISTS klines_view AS
SELECT
    pool_id,
    interval_start,
    interval,
    argMinMerge(open_price) as open,
    maxMerge(max_price) as high,
    minMerge(min_price) as low,
    argMaxMerge(close_price) as close,
    sumMerge(volume) as volume,
    sumMerge(quote_volume) as quote_volume,
    countMerge(trade_count) as trades
FROM klines
GROUP BY
    pool_id,
    interval,
    interval_start;




-- 1min kline
CREATE MATERIALIZED VIEW IF NOT EXISTS klines_mv TO klines AS
SELECT
    pool_id,
    toStartOfInterval(time, INTERVAL 1 MINUTE) as interval_start,
    '1m' as interval,
    argMinState(price, time) as open_price,    
    maxState(price) as max_price,
    minState(price) as min_price,
    argMaxState(price, time) as close_price,     
    sumState(base_quantity) as volume,
    sumState(quote_quantity) as quote_volume,
    countState() as trade_count
FROM trades
GROUP BY
    pool_id,
    interval_start,
    interval;

	-- 5min kline
CREATE MATERIALIZED VIEW IF NOT EXISTS klines_5m_mv TO klines AS
SELECT
    pool_id,
    toStartOfInterval(time, INTERVAL 5 MINUTE) as interval_start,
    '5m' as interval,
    argMinState(price, time) as open_price,
    maxState(price) as max_price,
    minState(price) as min_price,
    argMaxState(price, time) as close_price,
    sumState(base_quantity) as volume,
    sumState(quote_quantity) as quote_volume,
    countState() as trade_count
FROM trades
GROUP BY
    pool_id,
    interval_start,
    interval;

-- 15min kline
CREATE MATERIALIZED VIEW IF NOT EXISTS klines_15m_mv TO klines AS
SELECT
    pool_id,
    toStartOfInterval(time, INTERVAL 15 MINUTE) as interval_start,
    '15m' as interval,
    argMinState(price, time) as open_price,
    maxState(price) as max_price,
    minState(price) as min_price,
    argMaxState(price, time) as close_price,
    sumState(base_quantity) as volume,
    sumState(quote_quantity) as quote_volume,
    countState() as trade_count
FROM trades
GROUP BY
    pool_id,
    interval_start,
    interval;

-- 30min kline
CREATE MATERIALIZED VIEW IF NOT EXISTS klines_30m_mv TO klines AS
SELECT
    pool_id,
    toStartOfInterval(time, INTERVAL 30 MINUTE) as interval_start,
    '30m' as interval,
    argMinState(price, time) as open_price,
    maxState(price) as max_price,
    minState(price) as min_price,
    argMaxState(price, time) as close_price,
    sumState(base_quantity) as volume,
    sumState(quote_quantity) as quote_volume,
    countState() as trade_count
FROM trades
GROUP BY
    pool_id,
    interval_start,
    interval;

-- 1hour kline
CREATE MATERIALIZED VIEW IF NOT EXISTS klines_1h_mv TO klines AS
SELECT
    pool_id,
    toStartOfInterval(time, INTERVAL 1 HOUR) as interval_start,
    '1h' as interval,
    argMinState(price, time) as open_price,
    maxState(price) as max_price,
    minState(price) as min_price,
    argMaxState(price, time) as close_price,
    sumState(base_quantity) as volume,
    sumState(quote_quantity) as quote_volume,
    countState() as trade_count
FROM trades
GROUP BY
    pool_id,
    interval_start,
    interval;

-- 4hour kline
CREATE MATERIALIZED VIEW IF NOT EXISTS klines_4h_mv TO klines AS
SELECT
    pool_id,
    toStartOfInterval(time, INTERVAL 4 HOUR) as interval_start,
    '4h' as interval,
    argMinState(price, time) as open_price,
    maxState(price) as max_price,
    minState(price) as min_price,
    argMaxState(price, time) as close_price,
    sumState(base_quantity) as volume,
    sumState(quote_quantity) as quote_volume,
    countState() as trade_count
FROM trades
GROUP BY
    pool_id,
    interval_start,
    interval;

-- 1day kline
CREATE MATERIALIZED VIEW IF NOT EXISTS klines_1d_mv TO klines AS
SELECT
    pool_id,
    toStartOfDay(time) as interval_start,
    '1d' as interval,
    argMinState(price, time) as open_price,
    maxState(price) as max_price,
    minState(price) as min_price,
    argMaxState(price, time) as close_price,
    sumState(base_quantity) as volume,
    sumState(quote_quantity) as quote_volume,
    countState() as trade_count
FROM trades
GROUP BY
    pool_id,
    interval_start,
    interval;

-- 1week kline
CREATE MATERIALIZED VIEW IF NOT EXISTS klines_1w_mv TO klines AS
SELECT
    pool_id,
    toStartOfWeek(time) as interval_start,
    '1w' as interval,
    argMinState(price, time) as open_price,
    maxState(price) as max_price,
    minState(price) as min_price,
    argMaxState(price, time) as close_price,
    sumState(base_quantity) as volume,
    sumState(quote_quantity) as quote_volume,
    countState() as trade_count
FROM trades
GROUP BY
    pool_id,
    interval_start,
    interval;

-- 1month kline
CREATE MATERIALIZED VIEW IF NOT EXISTS klines_1M_mv TO klines AS
SELECT
    pool_id,
    toStartOfMonth(time) as interval_start,
    '1M' as interval,
    argMinState(price, time) as open_price,
    maxState(price) as max_price,
    minState(price) as min_price,
    argMaxState(price, time) as close_price,
    sumState(base_quantity) as volume,
    sumState(quote_quantity) as quote_volume,
    countState() as trade_count
FROM trades
GROUP BY
    pool_id,
    interval_start,
    interval;

