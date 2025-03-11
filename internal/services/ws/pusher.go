package ws

import (
	"context"
	"exapp-go/internal/entity"

	"github.com/zishang520/socket.io/v2/socket"
)

const (
	PushEventBalanceUpdate   = "balance_update"
	PushEventOrderUpdate     = "order_update"
	PushEventTradeUpdate     = "trade"
	PushEventDepthUpdate     = "depth"
	PushEventKlineUpdate     = "kline"
	PushEventPoolStatsUpdate = "pool_stats"
	PushEventUserCredential  = "new_user_credential"
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
		PoolID:    data.PoolID,
		Type:      SubTypeDepth,
		Precision: data.Precision,
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
func (p *Pusher) PushOrderUpdate(account string, update entity.Order) {
	// Push order update to specific user
	p.pushToUser(account, PushEventOrderUpdate, update)
}

// Push user credential
func (p *Pusher) PushUserCredential(account string, data entity.UserCredential) {
	p.pushToUser(account, PushEventUserCredential, data)
}

// Push pool stats data
func (p *Pusher) PushPoolStats(data entity.PoolStats) {
	sub := Subscription{
		PoolID: data.PoolID,
		Type:   SubTypePoolStats,
	}
	p.server.Broadcast(sub, PushEventPoolStatsUpdate, data)

	p.server.io.To(socket.Room("all_pool_stats")).Emit(PushEventPoolStatsUpdate, data)
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
