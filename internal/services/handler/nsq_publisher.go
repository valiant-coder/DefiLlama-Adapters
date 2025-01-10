package handler

import (
	"encoding/json"
	"exapp-go/internal/entity"
	"exapp-go/pkg/nsqutil"
	"log"
)

const (
	TopicCdexUpdates = "cdex_updates"
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
}

// NewNSQPublisher creates a new NSQ publisher
func NewNSQPublisher(nsqdAddrs []string) (*NSQPublisher, error) {
	publisher := nsqutil.NewPublisher(nsqdAddrs)
	return &NSQPublisher{
		publisher: publisher,
	}, nil
}

// Close closes the publisher connection
func (p *NSQPublisher) Close() {
	p.publisher.Stop()
}

// PublishOrderUpdate publishes an order update message
func (p *NSQPublisher) PublishOrderUpdate(account string, id string) error {
	msg := struct {
		Type string      `json:"type"`
		Data interface{} `json:"data"`
	}{
		Type: MsgTypeOrderUpdate,
		Data: OrderUpdate{
			Account: account,
			ID:      id,
		},
	}

	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Marshal order update failed: %v", err)
		return err
	}

	return p.publisher.Publish(TopicCdexUpdates, data)
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
	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Marshal balance update failed: %v", err)
		return err
	}

	return p.publisher.Publish(TopicCdexUpdates, data)
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

	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Marshal trade update failed: %v", err)
		return err
	}

	return p.publisher.Publish(TopicCdexUpdates, data)
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

	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Marshal depth update failed: %v", err)
		return err
	}

	return p.publisher.Publish(TopicCdexUpdates, data)
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

	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Marshal kline update failed: %v", err)
		return err
	}

	return p.publisher.Publish(TopicCdexUpdates, data)
}
