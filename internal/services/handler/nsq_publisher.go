package handler

import (
	"exapp-go/internal/entity"
	"exapp-go/pkg/hyperion"
	"exapp-go/pkg/nsqutil"
	"fmt"
	"log"
	"reflect"
	"sync"
	"time"
)

const (
	TopicCdexUpdates = "cdex_updates"
	cleanupInterval  = 24 * time.Hour // Cleanup interval set to 24 hours
)

// NSQ message types
const (
	MsgTypeOrderUpdate   = "order_update"
	MsgTypeBalanceUpdate = "balance_update"
	MsgTypeTradeUpdate   = "trade_update"
	MsgTypeDepthUpdate   = "depth_update"
	MsgTypeKlineUpdate   = "kline_update"
)

// OrderUpdate represents an order update message
type OrderUpdate struct {
	Account string `json:"account"`
	ID      string `json:"id"`
}

// NSQPublisher handles publishing messages to NSQ
type NSQPublisher struct {
	publisher *nsqutil.Client
	lastKline map[string]entity.Kline // key: symbol
	lastOrder map[string]OrderUpdate  // key: account-id
	mutex     sync.RWMutex
	stopChan  chan struct{} // Channel to stop cleanup goroutine
}

// NewNSQPublisher creates a new NSQ publisher
func NewNSQPublisher(nsqdAddrs []string) (*NSQPublisher, error) {
	publisher := nsqutil.NewPublisher(nsqdAddrs)
	p := &NSQPublisher{
		publisher: publisher,
		lastKline: make(map[string]entity.Kline),
		lastOrder: make(map[string]OrderUpdate),
		mutex:     sync.RWMutex{},
		stopChan:  make(chan struct{}),
	}

	// Start cleanup goroutine
	go p.startCleanupRoutine()

	return p, nil
}

// Close closes the publisher connection
func (p *NSQPublisher) Close() {
	close(p.stopChan) // Send stop signal
	p.publisher.Stop()
}

// startCleanupRoutine starts periodic cleanup routine
func (p *NSQPublisher) startCleanupRoutine() {
	ticker := time.NewTicker(cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			p.cleanup()
		case <-p.stopChan:
			return
		}
	}
}

// cleanup cleans expired cache data
func (p *NSQPublisher) cleanup() {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	// Clean lastKline and lastOrder maps
	p.lastKline = make(map[string]entity.Kline)
	p.lastOrder = make(map[string]OrderUpdate)

	log.Printf("NSQPublisher cache cleaned up")
}

// PublishOrderUpdate publishes an order update message
func (p *NSQPublisher) PublishOrderUpdate(account string, id string) error {
	orderUpdate := OrderUpdate{
		Account: account,
		ID:      id,
	}

	p.mutex.RLock()
	key := fmt.Sprintf("%s-%s", account, id)
	lastOrder, exists := p.lastOrder[key]
	p.mutex.RUnlock()

	if exists && lastOrder == orderUpdate {
		log.Printf("skip duplicate order update: %s-%s", account, id)
		return nil
	}

	log.Printf("publish order update: %s-%s", account, id)
	msg := struct {
		Type string      `json:"type"`
		Data interface{} `json:"data"`
	}{
		Type: MsgTypeOrderUpdate,
		Data: orderUpdate,
	}

	err := p.publisher.Publish(TopicCdexUpdates, msg)
	if err == nil {
		p.mutex.Lock()
		p.lastOrder[key] = orderUpdate
		p.mutex.Unlock()
	}
	return err
}

// PublishBalanceUpdate publishes a balance update message
func (p *NSQPublisher) PublishBalanceUpdate(account string) error {
	msg := struct {
		Type string      `json:"type"`
		Data interface{} `json:"data"`
	}{
		Type: MsgTypeBalanceUpdate,
		Data: account,
	}

	return p.publisher.Publish(TopicCdexUpdates, msg)
}

// PublishTradeUpdate publishes a trade update message
func (p *NSQPublisher) PublishTradeUpdate(trade entity.Trade) error {
	msg := struct {
		Type string      `json:"type"`
		Data interface{} `json:"data"`
	}{
		Type: MsgTypeTradeUpdate,
		Data: trade,
	}

	return p.publisher.Publish(TopicCdexUpdates, msg)
}

// PublishDepthUpdate publishes a depth update message
func (p *NSQPublisher) PublishDepthUpdate(depth entity.Depth) error {
	msg := struct {
		Type string      `json:"type"`
		Data interface{} `json:"data"`
	}{
		Type: MsgTypeDepthUpdate,
		Data: depth,
	}

	return p.publisher.Publish(TopicCdexUpdates, msg)
}

// PublishKlineUpdate publishes a kline update message
func (p *NSQPublisher) PublishKlineUpdate(kline entity.Kline) error {
	p.mutex.RLock()
	symbol := fmt.Sprintf("%d-%s-%d-%v", kline.PoolID, kline.Interval, kline.Timestamp.Timestamp(), kline.Turnover)
	lastKline, exists := p.lastKline[symbol]
	p.mutex.RUnlock()

	if exists && reflect.DeepEqual(lastKline, kline) {
		log.Printf("skip duplicate kline update for symbol: %s", symbol)
		return nil
	}

	msg := struct {
		Type string      `json:"type"`
		Data interface{} `json:"data"`
	}{
		Type: MsgTypeKlineUpdate,
		Data: kline,
	}

	err := p.publisher.Publish(TopicCdexUpdates, msg)
	if err == nil {
		p.mutex.Lock()
		p.lastKline[symbol] = kline
		p.mutex.Unlock()
	}
	return err
}

func (p *NSQPublisher) DeferPublishCreateOrder(action hyperion.Action) error {
	p.publisher.DeferredPublish(TopicActionSync, 1*time.Second, action)
	return nil
}
