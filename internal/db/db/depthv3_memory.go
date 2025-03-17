package db

import (
	"context"
	"sort"
	"sync"

	"github.com/shopspring/decimal"
)

// MemoryDepth is a structure for storing depth data in memory
type MemoryDepth struct {
	sync.RWMutex
	// Store processed order IDs
	processedIDs map[uint64]map[string]struct{}
	// Store buy/sell order data at different precisions
	// poolID -> side -> precision -> price -> amount
	depthData map[uint64]map[string]map[string]map[string]decimal.Decimal
	// Maintain sorted price lists
	// poolID -> side -> precision -> sorted prices
	sortedPrices map[uint64]map[string]map[string][]string
}

// NewMemoryDepth creates a new instance of memory depth storage
func NewMemoryDepth() *MemoryDepth {
	return &MemoryDepth{
		processedIDs: make(map[uint64]map[string]struct{}),
		depthData:    make(map[uint64]map[string]map[string]map[string]decimal.Decimal),
		sortedPrices: make(map[uint64]map[string]map[string][]string),
	}
}

// Ensure pool and side maps exist
func (m *MemoryDepth) ensurePoolAndSideMaps(poolID uint64, side string) {
	if _, ok := m.depthData[poolID]; !ok {
		m.depthData[poolID] = make(map[string]map[string]map[string]decimal.Decimal)
		m.sortedPrices[poolID] = make(map[string]map[string][]string)
	}
	if _, ok := m.depthData[poolID][side]; !ok {
		m.depthData[poolID][side] = make(map[string]map[string]decimal.Decimal)
		m.sortedPrices[poolID][side] = make(map[string][]string)
	}
}

// Ensure precision map exists
func (m *MemoryDepth) ensurePrecisionMap(poolID uint64, side, precision string) {
	if _, ok := m.depthData[poolID][side][precision]; !ok {
		m.depthData[poolID][side][precision] = make(map[string]decimal.Decimal)
		m.sortedPrices[poolID][side][precision] = make([]string, 0)
	}
}

// Update sorted price list
func (m *MemoryDepth) updateSortedPrices(poolID uint64, side, precision string) {
	prices := make([]string, 0, len(m.depthData[poolID][side][precision]))
	for price := range m.depthData[poolID][side][precision] {
		if m.depthData[poolID][side][precision][price].GreaterThan(decimal.NewFromFloat(0.00000001)) {
			prices = append(prices, price)
		}
	}

	// Sort based on buy/sell direction
	sort.Slice(prices, func(i, j int) bool {
		priceI := decimal.RequireFromString(prices[i])
		priceJ := decimal.RequireFromString(prices[j])
		if side == "buy" {
			return priceI.GreaterThan(priceJ)
		}
		return priceI.LessThan(priceJ)
	})

	m.sortedPrices[poolID][side][precision] = prices
}

// UpdateDepthV3 updates depth data
func (m *MemoryDepth) UpdateDepthV3(ctx context.Context, params []UpdateDepthParams) ([]DepthChange, error) {
	m.Lock()
	defer m.Unlock()

	// Check UniqID
	for _, param := range params {
		if param.UniqID != "" {
			if _, ok := m.processedIDs[param.PoolID]; !ok {
				m.processedIDs[param.PoolID] = make(map[string]struct{})
			}
			if _, exists := m.processedIDs[param.PoolID][param.UniqID]; exists {
				return nil, nil
			}
		}
	}

	// Add UniqID to processed set
	for _, param := range params {
		if param.UniqID != "" {
			m.processedIDs[param.PoolID][param.UniqID] = struct{}{}
		}
	}

	params = aggregateParams(params)
	var changes []DepthChange

	for _, param := range params {
		side := "sell"
		if param.IsBuy {
			side = "buy"
		}

		precisions, slots := calculateAllSlots(param.Price, param.IsBuy)
		for i, precision := range precisions {
			slot := slots[i]
			m.ensurePoolAndSideMaps(param.PoolID, side)
			m.ensurePrecisionMap(param.PoolID, side, precision)

			// Update amount
			currentAmount := m.depthData[param.PoolID][side][precision][slot]
			newAmount := currentAmount.Add(param.Amount)

			if newAmount.LessThan(decimal.NewFromFloat(0.00000001)) {
				delete(m.depthData[param.PoolID][side][precision], slot)
			} else {
				m.depthData[param.PoolID][side][precision][slot] = newAmount
			}

			// Update sorted price list
			m.updateSortedPrices(param.PoolID, side, precision)

			// Record changes
			fixedNewAmount := newAmount.Truncate(8)
			if fixedNewAmount.LessThan(decimal.Zero) {
				fixedNewAmount = decimal.Zero
			}

			changes = append(changes, DepthChange{
				PoolID:    param.PoolID,
				IsBuy:     param.IsBuy,
				Price:     decimal.RequireFromString(slot),
				Amount:    fixedNewAmount,
				Precision: precision,
			})
		}
	}

	return changes, nil
}

// GetDepthV3 gets depth data for specified precision
func (m *MemoryDepth) GetDepthV3(ctx context.Context, poolID uint64, precision string, limit int) (Depth, error) {
	m.RLock()
	defer m.RUnlock()

	depth := Depth{
		PoolID: poolID,
		Bids:   [][]string{},
		Asks:   [][]string{},
	}

	// Get buy orders
	if prices, ok := m.sortedPrices[poolID]["buy"][precision]; ok {
		for _, price := range prices {
			if len(depth.Bids) >= limit {
				break
			}
			if amount, exists := m.depthData[poolID]["buy"][precision][price]; exists {
				quantity := amount.Truncate(8)
				if !quantity.IsZero() {
					depth.Bids = append(depth.Bids, []string{price, quantity.String()})
				}
			}
		}
	}

	// Get sell orders
	if prices, ok := m.sortedPrices[poolID]["sell"][precision]; ok {
		for _, price := range prices {
			if len(depth.Asks) >= limit {
				break
			}
			if amount, exists := m.depthData[poolID]["sell"][precision][price]; exists {
				quantity := amount.Truncate(8)
				if !quantity.IsZero() {
					depth.Asks = append(depth.Asks, []string{price, quantity.String()})
				}
			}
		}
	}

	return depth, nil
}

// ClearDepthsV3 clears depth data
func (m *MemoryDepth) ClearDepthsV3(ctx context.Context, poolID uint64) error {
	m.Lock()
	defer m.Unlock()

	delete(m.processedIDs, poolID)
	delete(m.depthData, poolID)
	delete(m.sortedPrices, poolID)

	return nil
}

// CleanInvalidDepth cleans invalid depth data
func (m *MemoryDepth) CleanInvalidDepth(ctx context.Context, poolID uint64, lastPrice decimal.Decimal, isBuy bool) (int64, error) {
	m.Lock()
	defer m.Unlock()

	var totalCleaned int64

	for _, precision := range SupportedPrecisions {
		var side string
		var compareFunc func(price decimal.Decimal) bool

		if isBuy {
			side = "sell"
			precisionDecimal, _ := decimal.NewFromString(precision)
			slotPrice := lastPrice.Div(precisionDecimal).Ceil().Mul(precisionDecimal)
			compareFunc = func(price decimal.Decimal) bool {
				return price.LessThan(slotPrice)
			}
		} else {
			side = "buy"
			precisionDecimal, _ := decimal.NewFromString(precision)
			slotPrice := lastPrice.Div(precisionDecimal).Floor().Mul(precisionDecimal)
			compareFunc = func(price decimal.Decimal) bool {
				return price.GreaterThan(slotPrice)
			}
		}

		if _, ok := m.depthData[poolID][side][precision]; ok {
			var pricesToDelete []string
			for priceStr := range m.depthData[poolID][side][precision] {
				price := decimal.RequireFromString(priceStr)
				if compareFunc(price) {
					pricesToDelete = append(pricesToDelete, priceStr)
				}
			}

			for _, price := range pricesToDelete {
				delete(m.depthData[poolID][side][precision], price)
				totalCleaned++
			}

			if len(pricesToDelete) > 0 {
				m.updateSortedPrices(poolID, side, precision)
			}
		}
	}

	return totalCleaned, nil
}
