package ws

import (
	"context"
	"exapp-go/internal/entity"
)

const (
	PushEventBalanceUpdate = "balance_update"
	PushEventOrderUpdate   = "order_update"
	PushEventTradeUpdate   = "trade"
	PushEventDepthUpdate   = "depth"
	PushEventKlineUpdate   = "kline"
)

// Balance update
type BalanceUpdate struct {
	Account string `json:"account"`
}

// Order update
type OrderUpdate struct {
	Account string `json:"account"`
	ID      string `json:"id"`
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
func (p *Pusher) PushKline(data entity.Kline) {
	sub := Subscription{
		PoolID:   data.PoolID,
		Type:     SubTypeKline,
		Interval: data.Interval,
	}
	p.server.Broadcast(sub, PushEventKlineUpdate, data)
}

// Push depth data
func (p *Pusher) PushDepth(data entity.Depth) {
	sub := Subscription{
		PoolID: data.PoolID,
		Type:   SubTypeDepth,
	}
	p.server.Broadcast(sub, PushEventDepthUpdate, data)
}

// Push trade data
func (p *Pusher) PushTrade(data entity.Trade) {
	sub := Subscription{
		PoolID: data.PoolID,
		Type:   SubTypeTrades,
	}
	p.server.Broadcast(sub, PushEventTradeUpdate, data)
}

// Push balance update
func (p *Pusher) PushBalanceUpdate(account string) {
	// Push balance update to specific user
	p.pushToUser(account, PushEventBalanceUpdate, BalanceUpdate{Account: account})
}

// Push order update
func (p *Pusher) PushOrderUpdate(account string, update OrderUpdate) {
	// Push order update to specific user
	p.pushToUser(account, PushEventOrderUpdate, update)
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
