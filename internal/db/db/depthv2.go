package db

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
	"github.com/shopspring/decimal"
)

// UpdateDepthParams parameters for updating depth
type UpdateDepthParams struct {
	PoolID uint64
	UniqID string
	IsBuy  bool
	Price  decimal.Decimal
	// Positive means add, negative means subtract
	Amount decimal.Decimal
}

type Depth struct {
	PoolID uint64
	Bids   [][]string
	Asks   [][]string
}

type DepthChange struct {
	PoolID    uint64
	IsBuy     bool
	Price     decimal.Decimal
	Amount    decimal.Decimal
	Precision string
}

// Ordered list of supported precisions (from small to large)
var SupportedPrecisions = []string{
	"0.00000001",
	"0.0000001",
	"0.000001",
	"0.00001",
	"0.0001",
	"0.001",
	"0.01",
	"0.1",
	"1",
	"10",
	"100",
	"1000",
	"10000",
}

// Calculate price slots for all precisions
func calculateAllSlots(price decimal.Decimal, isBid bool) map[string]string {
	slots := make(map[string]string)

	for _, p := range SupportedPrecisions {
		precision, _ := decimal.NewFromString(p)
		if isBid {
			slot := price.Div(precision).Floor().Mul(precision)
			slots[p] = slot.String()
		} else {
			slot := price.Div(precision).Ceil().Mul(precision)
			slots[p] = slot.String()
		}
	}
	return slots
}

// Update order (automatically updates all precision levels)
func (r *Repo) UpdateDepthV2(ctx context.Context, params []UpdateDepthParams) ([]DepthChange, error) {
	// Check UniqID
	for _, param := range params {
		if param.UniqID != "" {
			exists, err := r.redis.SIsMember(ctx, fmt.Sprintf("depth:%d:processed_ids", param.PoolID), param.UniqID).Result()
			if err != nil {
				return nil, fmt.Errorf("check uniq id error: %w", err)
			}
			if exists {
				return nil, nil
			}
		}
	}

	// Add UniqID to processed set
	pipe := r.redis.Pipeline()
	for _, param := range params {
		if param.UniqID != "" {
			pipe.SAdd(ctx, fmt.Sprintf("depth:%d:processed_ids", param.PoolID), param.UniqID)
		}
	}
	if _, err := pipe.Exec(ctx); err != nil {
		return nil, fmt.Errorf("add uniq id error: %w", err)
	}
	params = aggregateParams(params)

	script := redis.NewScript(`
	local hashKey = KEYS[1]
	local sortedSetKey = KEYS[2]
	local slot = ARGV[1]
	local amount = ARGV[2]

	local newTotal = redis.call('HINCRBYFLOAT', hashKey, slot, amount)
	
	if tonumber(newTotal) > 0 then
		local score = redis.call('ZSCORE', sortedSetKey, slot)
		if not score then
			redis.call('ZADD', sortedSetKey, slot, slot)
		end
	else
		redis.call('HDEL', hashKey, slot)
		redis.call('ZREM', sortedSetKey, slot)
	end

	return tostring(newTotal)
	`)

	var changes []DepthChange

	// Process each update parameter
	for _, param := range params {
		side := "sell"
		if param.IsBuy {
			side = "buy"
		}

		// Calculate slots for all precisions
		slots := calculateAllSlots(param.Price, param.IsBuy)

		for precision, slot := range slots {
			// Generate keys for each precision
			hashKey := fmt.Sprintf("depth:%d:%s:%s:hash", param.PoolID, side, precision)
			sortedSetKey := fmt.Sprintf("depth:%d:%s:%s:sorted_set", param.PoolID, side, precision)

			// Execute script
			newTotal, err := script.Run(ctx, r.redis, []string{hashKey, sortedSetKey}, slot, param.Amount.String()).Result()
			if err != nil {
				return nil, fmt.Errorf("failed to execute script for precision %s: %v", precision, err)
			}

			// Only record changes for default precision
			fixedNewTotal := decimal.RequireFromString(fmt.Sprint(newTotal)).Truncate(8)
			if fixedNewTotal.Equal(decimal.Zero) {
				r.redis.HDel(ctx, hashKey, slot)
				r.redis.ZRem(ctx, sortedSetKey, slot)
			}
			changes = append(changes, DepthChange{
				PoolID:    param.PoolID,
				IsBuy:     param.IsBuy,
				Price:     param.Price,
				Amount:    fixedNewTotal,
				Precision: precision,
			})
		}
	}

	return changes, nil
}

// Get depth data for specified precision
func (r *Repo) GetDepthV2(ctx context.Context, poolId uint64, precision string, limit int) (Depth, error) {
	depth := Depth{
		PoolID: poolId,
		Bids:   [][]string{},
		Asks:   [][]string{},
	}

	// Get depth data using minimum precision

	// Get bids (buy orders)
	bidsHash := fmt.Sprintf("depth:%d:buy:%s:hash", poolId, precision)
	bidsSortedSet := fmt.Sprintf("depth:%d:buy:%s:sorted_set", poolId, precision)

	// Get asks (sell orders)
	asksHash := fmt.Sprintf("depth:%d:sell:%s:hash", poolId, precision)
	asksSortedSet := fmt.Sprintf("depth:%d:sell:%s:sorted_set", poolId, precision)

	// Get bids (high to low, limit 100)
	bids, err := r.redis.ZRevRange(ctx, bidsSortedSet, 0, int64(limit-1)).Result()
	if err != nil && err != redis.Nil {
		return depth, err
	}
	if len(bids) > 0 {
		quantities, err := r.redis.HMGet(ctx, bidsHash, bids...).Result()
		if err != nil {
			return depth, err
		}
		for i, price := range bids {
			if quantities[i] != nil {
				quantity := decimal.RequireFromString(quantities[i].(string)).Truncate(8)
				if quantity.IsZero() {
					continue
				}
				depth.Bids = append(depth.Bids, []string{price, quantity.String()})
			}
		}
	}

	// Get asks (low to high, limit 100)
	asks, err := r.redis.ZRange(ctx, asksSortedSet, 0, int64(limit-1)).Result()
	if err != nil && err != redis.Nil {
		return depth, err
	}
	if len(asks) > 0 {
		quantities, err := r.redis.HMGet(ctx, asksHash, asks...).Result()
		if err != nil {
			return depth, err
		}
		for i, price := range asks {
			if quantities[i] != nil {
				quantity := decimal.RequireFromString(quantities[i].(string)).Truncate(8)
				if quantity.IsZero() {
					continue
				}
				depth.Asks = append(depth.Asks, []string{price, quantity.String()})
			}
		}
	}

	return depth, nil
}

// Helper function: check if slice contains element
func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

// ClearDepths clears depth data
func (s *Repo) ClearDepthsV2(ctx context.Context, poolID uint64) error {
	keys := []string{
		fmt.Sprintf("depth:%d:processed_ids", poolID),
	}
	for _, precision := range SupportedPrecisions {
		keys = append(keys, []string{
			fmt.Sprintf("depth:%d:buy:%s:hash", poolID, precision),
			fmt.Sprintf("depth:%d:buy:%s:sorted_set", poolID, precision),
			fmt.Sprintf("depth:%d:sell:%s:hash", poolID, precision),
			fmt.Sprintf("depth:%d:sell:%s:sorted_set", poolID, precision),
		}...)
	}
	return s.redis.Del(ctx, keys...).Err()
}
