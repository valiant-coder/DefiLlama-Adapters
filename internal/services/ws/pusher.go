package ws

import (
	"context"
	"time"
)

// Kline data
type KlineData struct {
	PoolID    uint64    `json:"pool_id"`
	Interval  string    `json:"interval"`
	OpenTime  time.Time `json:"open_time"`
	Open      string    `json:"open"`
	High      string    `json:"high"`
	Low       string    `json:"low"`
	Close     string    `json:"close"`
	Volume    string    `json:"volume"`
	CloseTime time.Time `json:"close_time"`
}

// Depth data
type DepthData struct {
	PoolID uint64     `json:"pool_id"`
	Bids   [][]string `json:"bids"` // [price, quantity]
	Asks   [][]string `json:"asks"` // [price, quantity]
}

// Trade data
type TradeData struct {
	PoolID   uint64    `json:"pool_id"`
	ID       int64     `json:"id"`
	Price    string    `json:"price"`
	Quantity string    `json:"quantity"`
	Side     string    `json:"side"` // buy/sell
	Time     time.Time `json:"time"`
}

// Balance update
type BalanceUpdate struct {
	UserID    int64  `json:"user_id"`
	Asset     string `json:"asset"`
	Available string `json:"available"`
	Frozen    string `json:"frozen"`
}

// Order update
type OrderUpdate struct {
	UserID    int64     `json:"user_id"`
	OrderID   int64     `json:"order_id"`
	PoolID    uint64    `json:"pool_id"`
	Side      string    `json:"side"`
	Type      string    `json:"type"`
	Status    string    `json:"status"`
	Price     string    `json:"price"`
	Quantity  string    `json:"quantity"`
	Executed  string    `json:"executed"`
	Remaining string    `json:"remaining"`
	Time      time.Time `json:"time"`
}

// Push service
type Pusher struct {
	ctx    context.Context
	cancel context.CancelFunc
	server *Server
}

// Create new push service
func NewPusher(ctx context.Context, server *Server) *Pusher {
	ctx, cancel := context.WithCancel(ctx)
	return &Pusher{
		ctx:    ctx,
		cancel: cancel,
		server: server,
	}
}

// Push kline data
func (p *Pusher) PushKline(data KlineData) {
	sub := Subscription{
		PoolID:   data.PoolID,
		Type:     SubTypeKline,
		Interval: data.Interval,
	}
	p.server.Broadcast(sub, "kline", data)
}

// Push depth data
func (p *Pusher) PushDepth(data DepthData) {
	sub := Subscription{
		PoolID: data.PoolID,
		Type:   SubTypeDepth,
	}
	p.server.Broadcast(sub, "depth", data)
}

// Push trade data
func (p *Pusher) PushTrade(data TradeData) {
	sub := Subscription{
		PoolID: data.PoolID,
		Type:   SubTypeTrades,
	}
	p.server.Broadcast(sub, "trade", data)
}

// Push balance update
func (p *Pusher) PushBalanceUpdate(account string, update BalanceUpdate) {
	// Push balance update to specific user
	event := "balance_update"
	p.pushToUser(account, event, update)
}

// Push order update
func (p *Pusher) PushOrderUpdate(account string, update OrderUpdate) {
	// Push order update to specific user
	event := "order_update"
	p.pushToUser(account, event, update)
}

// Push data to specific user
func (p *Pusher) pushToUser(account string, event string, data interface{}) {
	// TODO: Implement user-specific push logic
	// Need to maintain mapping of user ID to Socket connection
	p.server.PushToAccount(account, event, data)
}

// Stop push service
func (p *Pusher) Stop() {
	p.cancel()
}
