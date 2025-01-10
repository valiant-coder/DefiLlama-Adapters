package db

import (
	"context"
	"errors"
	"exapp-go/pkg/utils"
	"fmt"
	"strings"

	"github.com/redis/go-redis/v9"
	"github.com/shopspring/decimal"
)

type ActionType string

const (
	ActionAdd ActionType = "add"
	ActionSub ActionType = "sub"
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

// Aggregate parameters
func aggregateParams(params []UpdateDepthParams) []UpdateDepthParams {
	var newParams []UpdateDepthParams
	var paramsMap = make(map[string]*UpdateDepthParams)
	for _, param := range params {
		key := fmt.Sprintf("%d:%s:%s", param.PoolID, map[bool]string{true: "1", false: "0"}[param.IsBuy], param.Price.String())
		if _, ok := paramsMap[key]; !ok {
			paramsMap[key] = &UpdateDepthParams{
				PoolID: param.PoolID,
				IsBuy:  param.IsBuy,
				Price:  param.Price,
				Amount: param.Amount,
				UniqID: param.UniqID,
			}
		} else {
			paramsMap[key].Amount = paramsMap[key].Amount.Add(param.Amount)
		}
	}
	for _, param := range paramsMap {
		newParams = append(newParams, *param)
	}
	return newParams
}

type DepthChange struct {
	PoolID uint64
	IsBuy  bool
	Price  decimal.Decimal
	Amount decimal.Decimal
}

// UpdateDepth updates depth data
func (s *Repo) UpdateDepth(ctx context.Context, params []UpdateDepthParams) ([]DepthChange, error) {
	for _, param := range params {
		if param.UniqID != "" {
			exists, err := s.IsMember(ctx, fmt.Sprintf("depth:%d:processed_ids", param.PoolID), param.UniqID)
			if err != nil {
				return nil, fmt.Errorf("check uniq id error: %w", err)
			}
			if exists {
				return nil, nil
			}
		}
	}

	keys := []string{}
	for _, param := range params {
		side := "asks"
		if param.IsBuy {
			side = "bids"
		}
		key := fmt.Sprintf("depth:%d:%s", param.PoolID, side)
		keys = append(keys, key)
	}
	// Remove duplicate keys
	keys = utils.RemoveDuplicate(keys)

	// Aggregate params
	params = aggregateParams(params)

	var changes []DepthChange

	err := s.Watch(ctx, func(tx *redis.Tx) error {
		pipe := tx.Pipeline()

		for _, param := range params {
			if param.UniqID != "" {
				pipe.SAdd(ctx, fmt.Sprintf("depth:%d:processed_ids", param.PoolID), param.UniqID)
			}
		}

		for _, param := range params {
			side := "asks"
			if param.IsBuy {
				side = "bids"
			}
			key := fmt.Sprintf("depth:%d:%s", param.PoolID, side)
			priceStr := param.Price.String()
			// Get current amount at this price level
			var existingAmount decimal.Decimal
			result, err := tx.ZRangeByScore(ctx, key, &redis.ZRangeBy{
				Min: priceStr,
				Max: priceStr,
			}).Result()

			if err != nil {
				return err
			}
			// If existing data found, parse the amount
			if len(result) > 0 {
				existingAmount = decimal.RequireFromString((strings.Split(result[0], ":")[1]))
			}

			// Calculate new amount
			var newAmount decimal.Decimal
			newAmount = existingAmount.Add(param.Amount)
			if newAmount.LessThan(decimal.Zero) {
				return fmt.Errorf("insufficient amount at price %s, existing %s, trying to reduce %s",
					param.Price.String(), existingAmount.String(), param.Amount.String())
			}

			// append depth changes
			changes = append(changes, DepthChange{
				PoolID: param.PoolID,
				IsBuy:  param.IsBuy,
				Price:  param.Price,
				Amount: newAmount,
			})

			if len(result) > 0 {
				member := result[0]
				pipe.ZRem(ctx, key, member)
			}

			if newAmount.GreaterThan(decimal.Zero) {
				// Update depth, member format: "price:amount"
				member := fmt.Sprintf("%s:%s", priceStr, newAmount.String())
				pipe.ZAdd(ctx, key, redis.Z{
					Score:  param.Price.InexactFloat64(), // Use price as score for sorting
					Member: member,                       // Store data in "price:amount" format
				})
			}

		}
		_, err := pipe.Exec(ctx)
		return err
	}, keys...)

	return changes, err
}

// GetDepth retrieves depth data
func (s *Repo) GetDepth(ctx context.Context, poolId uint64) (Depth, error) {
	depth := Depth{
		PoolID: poolId,
		Bids:   [][]string{},
		Asks:   [][]string{},
	}

	bidsKey := fmt.Sprintf("depth:%d:bids", poolId)
	asksKey := fmt.Sprintf("depth:%d:asks", poolId)

	// Bids from high to low
	bidsResult, err := s.rdb.single.ZRevRange(ctx, bidsKey, 0, -1).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return depth, nil
		}
		return depth, err
	}
	// Asks from low to high
	asksResult, err := s.rdb.single.ZRange(ctx, asksKey, 0, -1).Result()
	if err != nil {
		return depth, err
	}

	depth.Bids, err = parseDepth(bidsResult)
	if err != nil {
		return depth, err
	}
	depth.Asks, err = parseDepth(asksResult)
	if err != nil {
		return depth, err
	}
	return depth, nil
}

func parseDepth(result []string) ([][]string, error) {
	var depths [][]string
	for _, z := range result {
		parts := strings.Split(z, ":")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid depth data format: %v", z)
		}

		depths = append(depths, []string{parts[0], parts[1]})
	}

	return depths, nil
}

// ClearDepths clears depth data
func (s *Repo) ClearDepths(ctx context.Context, poolID uint64) error {
	keys := []string{
		fmt.Sprintf("depth:%d:bids", poolID),
		fmt.Sprintf("depth:%d:asks", poolID),
		fmt.Sprintf("depth:%d:processed_ids", poolID),
	}
	return s.CacheDel(ctx, keys...)
}
