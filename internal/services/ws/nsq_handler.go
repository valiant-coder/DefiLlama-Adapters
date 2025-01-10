package ws

import (
	"encoding/json"
	"exapp-go/internal/entity"
	"log"

	"github.com/nsqio/go-nsq"
)

// NSQ message types
const (
	MsgTypeOrderUpdate   = "order_update"
	MsgTypeBalanceUpdate = "balance_update"
	MsgTypeTradeUpdate   = "trade_update"
	MsgTypeDepthUpdate   = "depth_update"
	MsgTypeKlineUpdate   = "kline_update"
)

// Base NSQ message structure
type NSQMessage struct {
	Type    string          `json:"type"`
	Data    json.RawMessage `json:"data"`
}

// Handle NSQ message
func (s *Server) handleNSQMessage(msg *nsq.Message) error {
	var nsqMsg NSQMessage
	if err := json.Unmarshal(msg.Body, &nsqMsg); err != nil {
		log.Printf("Failed to unmarshal NSQ message: %v", err)
		return nil // Return nil to confirm message
	}

	switch nsqMsg.Type {
	case MsgTypeOrderUpdate:
		var orderUpdate OrderUpdate
		if err := json.Unmarshal(nsqMsg.Data, &orderUpdate); err != nil {
			log.Printf("Failed to unmarshal order update: %v", err)
			return nil
		}
		// Push order update to specific user
		s.pusher.PushOrderUpdate(orderUpdate.Account, orderUpdate)

	case MsgTypeTradeUpdate:
		var tradeData entity.Trade
		if err := json.Unmarshal(nsqMsg.Data, &tradeData); err != nil {
			log.Printf("Failed to unmarshal trade data: %v", err)
			return nil
		}
		// Broadcast trade update
		s.pusher.PushTrade(tradeData)

	case MsgTypeDepthUpdate:
		var depthData entity.Depth
		if err := json.Unmarshal(nsqMsg.Data, &depthData); err != nil {
			log.Printf("Failed to unmarshal depth data: %v", err)
			return nil
		}
		// Broadcast depth update
		s.pusher.PushDepth(depthData)

	case MsgTypeKlineUpdate:
		var klineData entity.Kline
		if err := json.Unmarshal(nsqMsg.Data, &klineData); err != nil {
			log.Printf("Failed to unmarshal kline data: %v", err)
			return nil
		}
		// Broadcast kline update
		s.pusher.PushKline(klineData)

	default:
		log.Printf("Unknown message type: %s", nsqMsg.Type)
	}

	return nil
}
