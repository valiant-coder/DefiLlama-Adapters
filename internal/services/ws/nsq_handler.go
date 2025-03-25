package ws

import (
	"encoding/json"
	"exapp-go/internal/entity"
	"fmt"
	"log"

	"github.com/nsqio/go-nsq"
)

const (
	TopicCdexUpdates = "cdex_updates"
)

// NSQ message types
const (
	MsgTypeOrderUpdate     = "order_update"
	MsgTypeBalanceUpdate   = "balance_update"
	MsgTypeTradeUpdate     = "trade_update"
	MsgTypeDepthUpdate     = "depth_update"
	MsgTypeKlineUpdate     = "kline_update"
	MsgTypePoolStatsUpdate = "pool_stats_update"
	MsgTypeUserCredential  = "new_user_credential"
)

// Base NSQ message structure
type NSQMessage struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

// Handle NSQ message
func (s *Server) handleNSQMessage(msg *nsq.Message) error {
	var nsqMsg NSQMessage
	if err := json.Unmarshal(msg.Body, &nsqMsg); err != nil {
		log.Printf("Failed to unmarshal NSQ message: %v,%v", err, string(msg.Body))
		return nil // Return nil to confirm message
	}

	switch nsqMsg.Type {
	case MsgTypeOrderUpdate:
		var order entity.Order
		if err := json.Unmarshal(nsqMsg.Data, &order); err != nil {
			log.Printf("Failed to unmarshal order update: %v", err)
			return nil
		}
		// Push order update to specific user
		s.pusher.PushOrderUpdate(fmt.Sprintf("%s@%s", order.Trader, order.Permission), order)

	case MsgTypeBalanceUpdate:
		var account string
		if err := json.Unmarshal(nsqMsg.Data, &account); err != nil {
			log.Printf("Failed to unmarshal balance update: %v", err)
			return nil
		}
		// Push balance update to specific user
		s.pusher.PushBalanceUpdate(account)

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

	case MsgTypePoolStatsUpdate:
		var poolStatsData entity.PoolStats
		if err := json.Unmarshal(nsqMsg.Data, &poolStatsData); err != nil {
			log.Printf("Failed to unmarshal pool stats data: %v", err)
			return nil
		}
		// Broadcast pool stats update
		s.pusher.PushPoolStats(poolStatsData)

	case MsgTypeUserCredential:
		var userCredential entity.UserCredential
		if err := json.Unmarshal(nsqMsg.Data, &userCredential); err != nil {
			log.Printf("Failed to unmarshal user credential: %v", err)
			return nil
		}
		s.pusher.PushUserCredential(userCredential.UID, userCredential)

	default:
		log.Printf("Unknown message type: %s", nsqMsg.Type)
	}

	return nil
}
