package db

import (
	"context"
	"fmt"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestNewMemoryDepth(t *testing.T) {
	md := NewMemoryDepth()
	assert.NotNil(t, md)
	assert.NotNil(t, md.processedIDs)
	assert.NotNil(t, md.depthData)
	assert.NotNil(t, md.sortedPrices)
}

func TestMemoryDepth_UpdateDepthV3(t *testing.T) {
	ctx := context.Background()
	md := NewMemoryDepth()

	tests := []struct {
		name        string
		params      []UpdateDepthParams
		wantChanges int
		wantErr     bool
	}{
		{
			name: "normal buy order update",
			params: []UpdateDepthParams{
				{
					PoolID: 1,
					UniqID: "order1",
					IsBuy:  true,
					Price:  decimal.NewFromFloat(100),
					Amount: decimal.NewFromFloat(1),
				},
			},
			wantChanges: 5, // Price 100 will generate changes for 5 precisions: 0.000000001, 0.01, 0.1, 1, 10
			wantErr:     false,
		},
		{
			name: "duplicate UniqID no update",
			params: []UpdateDepthParams{
				{
					PoolID: 1,
					UniqID: "order1",
					IsBuy:  true,
					Price:  decimal.NewFromFloat(100),
					Amount: decimal.NewFromFloat(1),
				},
			},
			wantChanges: 0,
			wantErr:     false,
		},
		{
			name: "sell order update",
			params: []UpdateDepthParams{
				{
					PoolID: 1,
					UniqID: "order2",
					IsBuy:  false,
					Price:  decimal.NewFromFloat(101),
					Amount: decimal.NewFromFloat(2),
				},
			},
			wantChanges: 5, // Price 101 will generate changes for 5 precisions
			wantErr:     false,
		},
		{
			name: "negative amount update",
			params: []UpdateDepthParams{
				{
					PoolID: 1,
					UniqID: "order3",
					IsBuy:  true,
					Price:  decimal.NewFromFloat(100),
					Amount: decimal.NewFromFloat(-0.5),
				},
			},
			wantChanges: 5, // Price 100 will generate changes for 5 precisions
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			changes, err := md.UpdateDepthV3(ctx, tt.params)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Len(t, changes, tt.wantChanges)

			// Additional validation of changes
			if len(changes) > 0 {
				// Verify amounts are correct
				for _, change := range changes {
					if tt.params[0].Amount.IsNegative() {
						assert.True(t, change.Amount.LessThanOrEqual(tt.params[0].Amount.Abs()))
					} else {
						assert.True(t, change.Amount.LessThanOrEqual(tt.params[0].Amount))
					}
				}
			}
		})
	}
}

func TestMemoryDepth_GetDepthV3(t *testing.T) {
	ctx := context.Background()
	md := NewMemoryDepth()

	// Prepare test data
	setupParams := []UpdateDepthParams{
		{
			PoolID: 1,
			UniqID: "order1",
			IsBuy:  true,
			Price:  decimal.NewFromFloat(100),
			Amount: decimal.NewFromFloat(1),
		},
		{
			PoolID: 1,
			UniqID: "order2",
			IsBuy:  false,
			Price:  decimal.NewFromFloat(101),
			Amount: decimal.NewFromFloat(2),
		},
	}

	_, err := md.UpdateDepthV3(ctx, setupParams)
	assert.NoError(t, err)

	tests := []struct {
		name      string
		poolID    uint64
		precision string
		limit     int
		wantBids  int
		wantAsks  int
	}{
		{
			name:      "get normal depth data",
			poolID:    1,
			precision: "0.01",
			limit:     10,
			wantBids:  1,
			wantAsks:  1,
		},
		{
			name:      "empty pool ID",
			poolID:    2,
			precision: "0.01",
			limit:     10,
			wantBids:  0,
			wantAsks:  0,
		},
		{
			name:      "limit records",
			poolID:    1,
			precision: "0.01",
			limit:     1,
			wantBids:  1,
			wantAsks:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			depth, err := md.GetDepthV3(ctx, tt.poolID, tt.precision, tt.limit)
			assert.NoError(t, err)
			assert.Len(t, depth.Bids, tt.wantBids)
			assert.Len(t, depth.Asks, tt.wantAsks)
		})
	}
}

func TestMemoryDepth_ClearDepthsV3(t *testing.T) {
	ctx := context.Background()
	md := NewMemoryDepth()

	// Prepare test data
	setupParams := []UpdateDepthParams{
		{
			PoolID: 1,
			UniqID: "order1",
			IsBuy:  true,
			Price:  decimal.NewFromFloat(100),
			Amount: decimal.NewFromFloat(1),
		},
	}

	_, err := md.UpdateDepthV3(ctx, setupParams)
	assert.NoError(t, err)

	// Clear data
	err = md.ClearDepthsV3(ctx, 1)
	assert.NoError(t, err)

	// Verify data has been cleared
	depth, err := md.GetDepthV3(ctx, 1, "0.01", 10)
	assert.NoError(t, err)
	assert.Empty(t, depth.Bids)
	assert.Empty(t, depth.Asks)
}

func TestMemoryDepth_CleanInvalidDepth(t *testing.T) {
	ctx := context.Background()
	md := NewMemoryDepth()

	// Prepare test data
	setupParams := []UpdateDepthParams{
		{
			PoolID: 1,
			UniqID: "order1",
			IsBuy:  true,
			Price:  decimal.NewFromFloat(100),
			Amount: decimal.NewFromFloat(1),
		},
		{
			PoolID: 1,
			UniqID: "order2",
			IsBuy:  false,
			Price:  decimal.NewFromFloat(101),
			Amount: decimal.NewFromFloat(2),
		},
		{
			PoolID: 1,
			UniqID: "order3",
			IsBuy:  false,
			Price:  decimal.NewFromFloat(102),
			Amount: decimal.NewFromFloat(3),
		},
	}

	_, err := md.UpdateDepthV3(ctx, setupParams)
	assert.NoError(t, err)

	// First verify initial state
	initialDepth, err := md.GetDepthV3(ctx, 1, "0.01", 10)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(initialDepth.Bids), "incorrect initial buy orders count")
	assert.Equal(t, 2, len(initialDepth.Asks), "incorrect initial sell orders count")

	tests := []struct {
		name        string
		poolID      uint64
		lastPrice   decimal.Decimal
		isBuy       bool
		wantCleaned int64
		wantBids    int
		wantAsks    int
	}{
		{
			name:        "clean buy orders",
			poolID:      1,
			lastPrice:   decimal.NewFromFloat(99),
			isBuy:       false,
			wantCleaned: 5, // number of precision levels cleaned for buy orders
			wantBids:    0, // buy orders should be completely cleaned
			wantAsks:    2, // sell orders should remain unchanged
		},
		{
			name:        "clean sell orders",
			poolID:      1,
			lastPrice:   decimal.NewFromFloat(101.5),
			isBuy:       true,
			wantCleaned: 5, // number of precision levels cleaned for one sell order
			wantBids:    0, // buy orders were cleaned in previous test
			wantAsks:    1, // should have one sell order remaining
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleaned, err := md.CleanInvalidDepth(ctx, tt.poolID, tt.lastPrice, tt.isBuy)
			assert.NoError(t, err)
			assert.Equal(t, tt.wantCleaned, cleaned, "cleaned count mismatch")

			// Verify remaining data
			depth, err := md.GetDepthV3(ctx, tt.poolID, "0.01", 10)
			assert.NoError(t, err)
			assert.Equal(t, tt.wantBids, len(depth.Bids), "remaining buy orders count mismatch")
			assert.Equal(t, tt.wantAsks, len(depth.Asks), "remaining sell orders count mismatch")

			// Verify price ranges
			if tt.isBuy {
				// Verify all sell orders have prices greater than lastPrice
				for _, ask := range depth.Asks {
					price := decimal.RequireFromString(ask[0])
					assert.True(t, price.GreaterThan(tt.lastPrice), "found sell order with price <= lastPrice")
				}
			} else {
				// Verify all buy orders have prices less than lastPrice
				for _, bid := range depth.Bids {
					price := decimal.RequireFromString(bid[0])
					assert.True(t, price.LessThan(tt.lastPrice), "found buy order with price >= lastPrice")
				}
			}
		})
	}
}

func TestMemoryDepth_Concurrency(t *testing.T) {
	ctx := context.Background()
	md := NewMemoryDepth()
	done := make(chan bool)

	// Concurrent updates
	go func() {
		for i := 0; i < 100; i++ {
			params := []UpdateDepthParams{
				{
					PoolID: 1,
					UniqID: fmt.Sprintf("order%d", i),
					IsBuy:  true,
					Price:  decimal.NewFromFloat(float64(100 + i)),
					Amount: decimal.NewFromFloat(1),
				},
			}
			_, err := md.UpdateDepthV3(ctx, params)
			assert.NoError(t, err)
		}
		done <- true
	}()

	// Concurrent reads
	go func() {
		for i := 0; i < 100; i++ {
			_, err := md.GetDepthV3(ctx, 1, "0.01", 10)
			assert.NoError(t, err)
		}
		done <- true
	}()

	// Wait for all operations to complete
	<-done
	<-done

	// Verify final state
	depth, err := md.GetDepthV3(ctx, 1, "0.01", 200)
	assert.NoError(t, err)
	assert.NotEmpty(t, depth.Bids)
}
