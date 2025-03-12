package handler

import (
	"exapp-go/internal/entity"
	"exapp-go/pkg/hyperion"
	"exapp-go/pkg/nsqutil"
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
	publisher *nsqutil.Publisher
}

// NewNSQPublisher creates a new NSQ publisher
func NewNSQPublisher(nsqdAddrs []string) (*NSQPublisher, error) {
	publisher := nsqutil.NewPublisher(nsqdAddrs)
	p := &NSQPublisher{
		publisher: publisher,
	}

	return p, nil
}

// Close closes the publisher connection
func (p *NSQPublisher) Close() {
	p.publisher.Stop()
}

// PublishOrderUpdate publishes an order update message
func (p *NSQPublisher) PublishOrderUpdate(account string, order entity.Order) error {
	msg := struct {
		Type string      `json:"type"`
		Data interface{} `json:"data"`
	}{
		Type: MsgTypeOrderUpdate,
		Data: order,
	}
	return p.publisher.Publish(TopicCdexUpdates, msg)

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
	msg := struct {
		Type string      `json:"type"`
		Data interface{} `json:"data"`
	}{
		Type: MsgTypeKlineUpdate,
		Data: kline,
	}

	return p.publisher.Publish(TopicCdexUpdates, msg)
}

func (p *NSQPublisher) DeferPublishCreateOrder(action hyperion.Action) error {
	p.publisher.DeferredPublish(TopicActionSync, 1*time.Second, action)
	return nil
}
