package db

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
	"github.com/shopspring/decimal"
)

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
}




// Calculate price slots for all precisions
func calculateAllSlots(price decimal.Decimal) map[string]string {
	slots := make(map[string]string)

	for _, p := range SupportedPrecisions {
		precision, _ := decimal.NewFromString(p)
		slot := price.Div(precision).Floor().Mul(precision)
		slots[p] = slot.String()
	}
	return slots
}

// Update order (automatically updates all precision levels)
func (r *Repo) UpdateDepthV2(ctx context.Context, symbol string, side string, price decimal.Decimal, amount decimal.Decimal) error {
	script := redis.NewScript(`
	local hashKey = KEYS[1]
	local sortedSetKey = KEYS[2]
	local slot = ARGV[1]
	local amount = ARGV[2]

	local newTotal = redis.call('HINCRBY', hashKey, slot, amount)
	
	if tonumber(newTotal) > 0 then
		local score = redis.call('ZSCORE', sortedSetKey, slot)
		if not score then
			redis.call('ZADD', sortedSetKey, slot, slot)
		end
	else
		redis.call('HDEL', hashKey, slot)
		redis.call('ZREM', sortedSetKey, slot)
	end

	return newTotal
	`)
	
	if side != "buy" && side != "sell" {
		return fmt.Errorf("invalid side: %s", side)
	}

	// Calculate slots for all precisions
	slots := calculateAllSlots(price)

	for precision, slot := range slots {
		// Generate keys for each precision
		hashKey := fmt.Sprintf("depth:%s:%s:%s:hash", symbol, side, precision)
		sortedSetKey := fmt.Sprintf("depth:%s:%s:%s:sorted_set", symbol, side, precision)

		// Execute script
		_, err := script.Run(ctx, r.redis, []string{hashKey, sortedSetKey}, slot, amount.String()).Result()
		if err != nil {
			return fmt.Errorf("failed to execute script for precision %s: %v", precision, err)
		}
	}

	return nil
}

// Get depth data for specified precision
func (r *Repo) GetDepthV2(ctx context.Context,symbol string, side string, precision string, limit int) (map[decimal.Decimal]decimal.Decimal, error) {
	if side != "buy" && side != "sell" {
		return nil, fmt.Errorf("invalid side: %s", side)
	}

	// Validate precision is supported
	if !contains(SupportedPrecisions, precision) {
		return nil, fmt.Errorf("unsupported precision: %s", precision)
	}

	// Generate Redis keys
	hashKey := fmt.Sprintf("depth:%s:%s:%s:hash", symbol, side, precision)
	sortedSetKey := fmt.Sprintf("depth:%s:%s:%s:sorted_set", symbol, side, precision)

	// Get sort direction
	var pricesCmd *redis.StringSliceCmd
	if side == "buy" {
		pricesCmd = r.redis.ZRevRange(ctx, sortedSetKey, 0, int64(limit-1))
	} else {
		pricesCmd = r.redis.ZRange(ctx, sortedSetKey, 0, int64(limit-1))
	}

	prices, err := pricesCmd.Result()
	if err != nil {
		return nil, err
	}

	// Batch get quantities
	quantities, err := r.redis.HMGet(ctx, hashKey, prices...).Result()
	if err != nil {
		return nil, err
	}

	// Convert results
	result := make(map[decimal.Decimal]decimal.Decimal)
	for i, p := range prices {
		price, _ := decimal.NewFromString(p)
		quantity, _ := decimal.NewFromString(quantities[i].(string))
		result[price] = quantity
	}

	return result, nil
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
