package db

import (
	"context"
	"fmt"
	"log"

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
	"0.000000001",
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
}

// getRelevantPrecisions returns precision levels based on price
func getRelevantPrecisions(price decimal.Decimal) []string {
	// Return empty slice if price is zero
	if price.IsZero() {
		return []string{}
	}

	// Predefined precision groups
	precisionGroups := map[string][]string{
		"high":   {"0.01", "0.1", "1", "10"},
		"medium": {"0.0001", "0.001", "0.01", "0.1"},
		"low":    {"0.000000001", "0.00000001", "0.0000001", "0.000001"},
	}

	// For prices greater than 10
	if price.GreaterThanOrEqual(decimal.NewFromInt(10)) {
		return precisionGroups["high"]
	}

	// For prices greater than 1
	if price.GreaterThanOrEqual(decimal.NewFromInt(1)) {
		return precisionGroups["medium"]
	}

	// For prices less than 1, find first non-zero decimal place
	decimalStr := price.String()
	firstNonZeroIndex := 2                           // Skip "0."
	for i := 2; i < len(decimalStr) && i < 11; i++ { // Limit search range to avoid large numbers
		if decimalStr[i] != '0' {
			firstNonZeroIndex = i
			break
		}
	}

	// Use lowest precision group if many decimal places
	if firstNonZeroIndex > 6 {
		return precisionGroups["low"]
	}

	// Select appropriate precisions based on first non-zero position
	var relevantPrecisions []string
	for _, p := range SupportedPrecisions {
		precision, _ := decimal.NewFromString(p)
		decimalPlaces := len(precision.String()) - 2 // Subtract "0."
		if decimalPlaces <= firstNonZeroIndex+3 && decimalPlaces >= firstNonZeroIndex-1 {
			relevantPrecisions = append(relevantPrecisions, p)
		}
	}

	// Keep maximum 4 precision levels
	if len(relevantPrecisions) > 4 {
		return relevantPrecisions[len(relevantPrecisions)-4:]
	}

	return relevantPrecisions
}

// Calculate price slots for all precisions
func calculateAllSlots(price decimal.Decimal, isBid bool) map[string]string {
	slots := make(map[string]string)

	// Get relevant precision levels
	relevantPrecisions := getRelevantPrecisions(price)
	if !contains(relevantPrecisions, "0.000000001") {
		relevantPrecisions = append(relevantPrecisions, "0.000000001")
	}

	for _, p := range relevantPrecisions {
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
	if len(params) == 0 {
		return nil, nil
	}

	// 1. Batch check UniqID
	pipe := r.redis.Pipeline()
	uniqIDChecks := make(map[string]*redis.BoolCmd)
	for _, param := range params {
		if param.UniqID != "" {
			key := fmt.Sprintf("{depth:%d}:processed_ids", param.PoolID)
			uniqIDChecks[param.UniqID] = pipe.SIsMember(ctx, key, param.UniqID)
		}
	}
	if _, err := pipe.Exec(ctx); err != nil {
		return nil, fmt.Errorf("check uniq ids error: %w", err)
	}

	// Filter already processed UniqIDs
	var validParams []UpdateDepthParams
	for _, param := range params {
		if param.UniqID != "" {
			if exists, err := uniqIDChecks[param.UniqID].Result(); err != nil || exists {
				continue
			}
		}
		validParams = append(validParams, param)
	}

	if len(validParams) == 0 {
		return nil, nil
	}

	// 2. Batch add UniqID to processed set
	pipe = r.redis.Pipeline()
	for _, param := range validParams {
		if param.UniqID != "" {
			pipe.SAdd(ctx, fmt.Sprintf("{depth:%d}:processed_ids", param.PoolID), param.UniqID)
		}
	}
	if _, err := pipe.Exec(ctx); err != nil {
		return nil, fmt.Errorf("add uniq ids error: %w", err)
	}

	// 3. Aggregate parameters
	validParams = aggregateParams(validParams)

	// 4. Optimized Redis script for batch updates
	script := redis.NewScript(`
	local updates = {}
	local n = 0
	for i = 1, #KEYS, 2 do
		local hashKey = KEYS[i]
		local sortedSetKey = KEYS[i + 1]
		local slot = ARGV[n * 2 + 1]
		local amount = ARGV[n * 2 + 2]
		
		local newTotal = redis.call('HINCRBYFLOAT', hashKey, slot, amount)
		
		if tonumber(newTotal) > 0 then
			if tonumber(newTotal) >= 0.00000001 then
				redis.call('ZADD', sortedSetKey, slot, slot)
			end
		else
			redis.call('HDEL', hashKey, slot)
			redis.call('ZREM', sortedSetKey, slot)
		end
		
		table.insert(updates, tostring(newTotal))
		n = n + 1
	end
	return updates
	`)

	var changes []DepthChange
	batchSize := 100 // Maximum number of items per batch

	// 5. Process updates in batches
	for i := 0; i < len(validParams); i += batchSize {
		end := i + batchSize
		if end > len(validParams) {
			end = len(validParams)
		}
		batchParams := validParams[i:end]

		var keys []string
		var args []interface{}

		// Prepare batch update parameters
		for _, param := range batchParams {
			side := "sell"
			if param.IsBuy {
				side = "buy"
			}

			// Calculate slots for all precisions
			slots := calculateAllSlots(param.Price, param.IsBuy)
			for precision, slot := range slots {
				hashKey := fmt.Sprintf("{depth:%d}:%s:%s:hash", param.PoolID, side, precision)
				sortedSetKey := fmt.Sprintf("{depth:%d}:%s:%s:sorted_set", param.PoolID, side, precision)

				keys = append(keys, hashKey, sortedSetKey)
				args = append(args, slot, param.Amount.String())
			}
		}

		// Execute batch update
		results, err := script.Run(ctx, r.redis, keys, args...).Result()
		if err != nil {
			return nil, fmt.Errorf("failed to execute batch update: %v", err)
		}

		// Process results
		resultList, ok := results.([]interface{})
		if !ok {
			return nil, fmt.Errorf("unexpected result type: %T", results)
		}

		// Record changes
		resultIdx := 0
		for _, param := range batchParams {
			slots := calculateAllSlots(param.Price, param.IsBuy)
			for precision, slot := range slots {
				if resultIdx >= len(resultList) {
					continue
				}

				if _, ok := resultList[resultIdx].(string); !ok {
					log.Printf("unexpected result type: %T", resultList[resultIdx])
					continue
				}
				newTotal := decimal.RequireFromString(resultList[resultIdx].(string)).Truncate(8)
				if newTotal.LessThan(decimal.Zero) {
					newTotal = decimal.Zero
				}

				changes = append(changes, DepthChange{
					PoolID:    param.PoolID,
					IsBuy:     param.IsBuy,
					Price:     decimal.RequireFromString(slot),
					Amount:    newTotal,
					Precision: precision,
				})

				resultIdx++
			}
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
	bidsHash := fmt.Sprintf("{depth:%d}:buy:%s:hash", poolId, precision)
	bidsSortedSet := fmt.Sprintf("{depth:%d}:buy:%s:sorted_set", poolId, precision)

	// Get asks (sell orders)
	asksHash := fmt.Sprintf("{depth:%d}:sell:%s:hash", poolId, precision)
	asksSortedSet := fmt.Sprintf("{depth:%d}:sell:%s:sorted_set", poolId, precision)

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
		fmt.Sprintf("{depth:%d}:processed_ids", poolID),
	}
	for _, precision := range SupportedPrecisions {
		keys = append(keys, []string{
			fmt.Sprintf("{depth:%d}:buy:%s:hash", poolID, precision),
			fmt.Sprintf("{depth:%d}:buy:%s:sorted_set", poolID, precision),
			fmt.Sprintf("{depth:%d}:sell:%s:hash", poolID, precision),
			fmt.Sprintf("{depth:%d}:sell:%s:sorted_set", poolID, precision),
		}...)
	}
	return s.redis.Del(ctx, keys...).Err()
}

func (r *Repo) CleanInvalidDepth(poolID uint64, lastPrice decimal.Decimal, isBuy bool) (int64, error) {
	ctx := context.Background()
	var totalCleaned int64

	pipe := r.redis.Pipeline()

	for _, precision := range SupportedPrecisions {
		var (
			hashKey      string
			sortedSetKey string
			min          string
			max          string
		)

		if isBuy {
			// For buy orders, clean all sell orders less than or equal to the executed price
			hashKey = fmt.Sprintf("{depth:%d}:sell:%s:hash", poolID, precision)
			sortedSetKey = fmt.Sprintf("{depth:%d}:sell:%s:sorted_set", poolID, precision)
			min = "-inf"
			// For sell orders, round up based on precision to ensure cleaning all orders below executed price (exclusive)
			precisionDecimal, _ := decimal.NewFromString(precision)
			slotPrice := lastPrice.Div(precisionDecimal).Ceil().Mul(precisionDecimal)
			max = "(" + slotPrice.String() // Use "(" prefix to exclude this price
		} else {
			// For sell orders, clean all buy orders greater than or equal to the executed price
			hashKey = fmt.Sprintf("{depth:%d}:buy:%s:hash", poolID, precision)
			sortedSetKey = fmt.Sprintf("{depth:%d}:buy:%s:sorted_set", poolID, precision)
			// For buy orders, round down based on precision to ensure cleaning all orders above executed price
			precisionDecimal, _ := decimal.NewFromString(precision)
			slotPrice := lastPrice.Div(precisionDecimal).Floor().Mul(precisionDecimal)
			min = "(" + slotPrice.String() // Use "(" prefix to exclude this price
			max = "+inf"
		}

		// Get list of prices to delete
		prices, err := r.redis.ZRangeByScore(ctx, sortedSetKey, &redis.ZRangeBy{
			Min: min,
			Max: max,
		}).Result()
		if err != nil {
			return 0, fmt.Errorf("get invalid prices error for precision %s: %w", precision, err)
		}

		if len(prices) > 0 {
			// Delete data from hash table
			pipe.HDel(ctx, hashKey, prices...)
			// Delete data from sorted set
			pipe.ZRemRangeByScore(ctx, sortedSetKey, min, max)
			totalCleaned += int64(len(prices))
		}
	}

	if totalCleaned > 0 {
		_, err := pipe.Exec(ctx)
		if err != nil {
			return 0, fmt.Errorf("clean invalid depth error: %w", err)
		}
	}

	return totalCleaned, nil
}
