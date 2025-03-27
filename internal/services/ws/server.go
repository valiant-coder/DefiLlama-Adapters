package ws

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"exapp-go/config"
	"exapp-go/internal/db/db"
	"exapp-go/internal/types"
	"exapp-go/pkg/nsqutil"

	"github.com/google/uuid"
	"github.com/spf13/cast"
	socketTypes "github.com/zishang520/engine.io/types"
	"github.com/zishang520/socket.io/v2/socket"
)

type SubscriptionType string

const (
	SubTypeKline     SubscriptionType = "kline"
	SubTypeDepth     SubscriptionType = "depth"
	SubTypeTrades    SubscriptionType = "trades"
	SubTypePoolStats SubscriptionType = "pool_stats"
)

type Subscription struct {
	PoolID    uint64
	Type      SubscriptionType
	Interval  string
	Precision string
}

type Server struct {
	io       *socket.Server
	pusher   *Pusher
	consumer *nsqutil.Consumer
}

// Create new WebSocket server
func NewServer(ctx context.Context) *Server {
	if ctx == nil {
		ctx = context.Background()
	}

	c := socket.DefaultServerOptions()
	c.SetServeClient(true)
	c.SetPingInterval(10 * time.Second)
	c.SetPingTimeout(5 * time.Second)
	c.SetMaxHttpBufferSize(1000000)
	c.SetConnectTimeout(5 * time.Second)
	c.SetCors(&socketTypes.Cors{
		Origin:      "*",
		Credentials: true,
	})

	io := socket.NewServer(nil, c)

	nsqCfg := config.Conf().Nsq
	server := &Server{
		io:       io,
		consumer: nsqutil.NewConsumer(nsqCfg.Lookupd, nsqCfg.LookupTTl),
	}
	server.pusher = NewPusher(ctx, server)

	// Setup NSQ message handlers
	server.setupNSQHandlers()

	// Set connection handler
	io.On("connection", server.handleConnection)

	return server
}

// Setup NSQ message handlers
func (s *Server) setupNSQHandlers() {
	if s.consumer != nil {
		s.consumer.Consume(string(types.TopicCdexUpdates), fmt.Sprintf("%s#ephemeral", uuid.New().String()), s.handleNSQMessage)
	}
}

// Get HTTP handler
func (s *Server) Handler() http.Handler {
	return s.io.ServeHandler(nil)
}

// Generate room name
func getRoomName(subType, poolID, interval, precision string) string {
	if poolID == "" {
		return ""
	}
	if interval != "" {
		return fmt.Sprintf("%s:%s:%s", subType, poolID, interval)
	}
	if precision != "" {
		return fmt.Sprintf("%s:%s:%s", subType, poolID, precision)
	}
	return fmt.Sprintf("%s:%s", subType, poolID)
}

// safeGetString safely extracts a string from an interface{}
func safeGetString(v interface{}) (string, bool) {
	if v == nil {
		return "", false
	}
	str, ok := v.(string)
	return str, ok
}

// safeGetFloat64 safely extracts a float64 from an interface{}
func safeGetFloat64(v interface{}) (float64, bool) {
	if v == nil {
		return 0, false
	}
	num, ok := v.(float64)
	return num, ok
}

// Handle new connection
func (s *Server) handleConnection(args ...interface{}) {
	if len(args) == 0 {
		return
	}

	client, ok := args[0].(*socket.Socket)
	if !ok || client == nil {
		return
	}

	// User authentication
	client.On("authenticate", func(args ...interface{}) {
		if len(args) < 1 {
			client.Emit("error", map[string]interface{}{
				"message": "Invalid authentication parameters",
			})
			return
		}

		account, ok := safeGetString(args[0])
		if !ok || account == "" {
			client.Emit("error", map[string]interface{}{
				"message": "Invalid account format",
			})
			return
		}
		uidAccount := strings.Split(account, "-")
		if len(uidAccount) > 2 || len(uidAccount) == 0 {
			client.Emit("error", map[string]interface{}{
				"message": "Invalid account format",
			})
			return
		}
		if len(uidAccount) == 2 {
			uid := uidAccount[0]
			eosAccount := uidAccount[1]
			uidRoom := socket.Room(fmt.Sprintf("account:%s", uid))
			client.Join(uidRoom)
			eosRoom := socket.Room(fmt.Sprintf("account:%s", eosAccount))
			client.Join(eosRoom)
		} else {
			accountRoom := socket.Room(fmt.Sprintf("account:%s", account))
			client.Join(accountRoom)
		}

		// Add user to dedicated room
		client.Emit("authenticated", map[string]interface{}{
			"status":  "success",
			"account": account,
		})
	})

	client.On("unauthenticate", func(args ...interface{}) {
		if len(args) < 1 {
			return
		}
		account, ok := safeGetString(args[0])
		if !ok || account == "" {
			return
		}
		uidAccount := strings.Split(account, "-")
		if len(uidAccount) > 2 || len(uidAccount) == 0 {
			return
		}
		if len(uidAccount) == 2 {
			uid := uidAccount[0]
			eosAccount := uidAccount[1]
			uidRoom := socket.Room(fmt.Sprintf("account:%s", uid))
			client.Leave(uidRoom)
			eosRoom := socket.Room(fmt.Sprintf("account:%s", eosAccount))
			client.Leave(eosRoom)
		} else {
			accountRoom := socket.Room(fmt.Sprintf("account:%s", account))
			client.Leave(accountRoom)
		}
	})

	// Subscribe to kline data
	client.On("subscribe_kline", func(args ...interface{}) {
		if len(args) < 2 {
			client.Emit("error", map[string]interface{}{
				"message": "Invalid kline subscription parameters",
			})
			return
		}

		poolIDFloat, ok := safeGetFloat64(args[0])
		if !ok {
			client.Emit("error", map[string]interface{}{
				"message": "Invalid pool_id format",
			})
			return
		}
		poolID := uint64(poolIDFloat)

		interval, ok := safeGetString(args[1])
		if !ok || interval == "" {
			client.Emit("error", map[string]interface{}{
				"message": "Invalid interval format",
			})
			return
		}

		room := socket.Room(getRoomName(string(SubTypeKline), cast.ToString(poolID), interval, ""))
		if room == "" {
			return
		}

		client.Join(room)
		client.Emit("subscribed", map[string]interface{}{
			"type":     SubTypeKline,
			"pool_id":  poolID,
			"interval": interval,
		})
	})

	// Subscribe to depth data
	client.On("subscribe_depth", func(args ...interface{}) {
		if len(args) < 1 {
			client.Emit("error", map[string]interface{}{
				"message": "Invalid depth subscription parameters",
			})
			return
		}

		poolIDFloat, ok := safeGetFloat64(args[0])
		if !ok {
			client.Emit("error", map[string]interface{}{
				"message": "Invalid pool_id format",
			})
			return
		}
		poolID := uint64(poolIDFloat)

		var precision string
		if len(args) == 1 {
			precision = "0.00000001"
		} else {
			precision, ok = safeGetString(args[1])
			if !ok {
				client.Emit("error", map[string]interface{}{
					"message": "Invalid precision format",
				})
				return
			}
		}

		validPrecision := false
		for _, p := range db.SupportedPrecisions {
			if p == precision {
				validPrecision = true
				break
			}
		}
		if !validPrecision {
			client.Emit("error", map[string]interface{}{
				"message": "Unsupported precision value",
			})
			return
		}

		room := socket.Room(getRoomName(string(SubTypeDepth), cast.ToString(poolID), "", precision))
		if room == "" {
			return
		}

		client.Join(room)
		client.Emit("subscribed", map[string]interface{}{
			"type":      SubTypeDepth,
			"pool_id":   poolID,
			"precision": precision,
		})
	})

	// Subscribe to trade data
	client.On("subscribe_trades", func(args ...interface{}) {
		if len(args) < 1 {
			client.Emit("error", map[string]interface{}{
				"message": "Invalid trades subscription parameters",
			})
			return
		}

		poolIDFloat, ok := safeGetFloat64(args[0])
		if !ok {
			client.Emit("error", map[string]interface{}{
				"message": "Invalid pool_id format",
			})
			return
		}
		poolID := uint64(poolIDFloat)

		room := socket.Room(getRoomName(string(SubTypeTrades), cast.ToString(poolID), "", ""))
		if room == "" {
			return
		}

		client.Join(room)
		client.Emit("subscribed", map[string]interface{}{
			"type":    SubTypeTrades,
			"pool_id": poolID,
		})
	})

	// Subscribe to all pool stats data
	client.On("subscribe_all_pool_stats", func(args ...interface{}) {
		room := socket.Room("all_pool_stats")
		client.Join(room)
		client.Emit("subscribed", map[string]interface{}{
			"type": "all_pool_stats",
		})
	})

	// Subscribe to pool stats data
	client.On("subscribe_pool_stats", func(args ...interface{}) {
		if len(args) < 1 {
			client.Emit("error", map[string]interface{}{
				"message": "Invalid pool stats subscription parameters",
			})
			return
		}

		poolIDFloat, ok := safeGetFloat64(args[0])
		if !ok {
			client.Emit("error", map[string]interface{}{
				"message": "Invalid pool_id format",
			})
			return
		}
		poolID := uint64(poolIDFloat)

		room := socket.Room(getRoomName(string(SubTypePoolStats), cast.ToString(poolID), "", ""))
		if room == "" {
			return
		}

		client.Join(room)
		client.Emit("subscribed", map[string]interface{}{
			"type":    SubTypePoolStats,
			"pool_id": poolID,
		})
	})

	// Unsubscribe
	client.On("unsubscribe", func(args ...interface{}) {
		if len(args) < 2 {
			client.Emit("error", map[string]interface{}{
				"message": "Invalid unsubscribe parameters",
			})
			return
		}

		subType, ok := safeGetString(args[0])
		if !ok {
			client.Emit("error", map[string]interface{}{
				"message": "Invalid subscription type",
			})
			return
		}

		poolIDFloat, ok := safeGetFloat64(args[1])
		if !ok {
			client.Emit("error", map[string]interface{}{
				"message": "Invalid pool_id format",
			})
			return
		}
		poolID := uint64(poolIDFloat)

		var room string
		if subType == string(SubTypeKline) {
			if len(args) < 3 {
				client.Emit("error", map[string]interface{}{
					"message": "Missing interval parameter for kline unsubscribe",
				})
				return
			}
			interval, ok := safeGetString(args[2])
			if !ok {
				client.Emit("error", map[string]interface{}{
					"message": "Invalid interval format",
				})
				return
			}
			room = getRoomName(subType, cast.ToString(poolID), interval, "")
		} else if subType == string(SubTypeDepth) {
			if len(args) < 2 {
				client.Emit("error", map[string]interface{}{
					"message": "Missing precision parameter for depth unsubscribe",
				})
				return
			}
			var precision string
			if len(args) == 2 {
				precision = "0.00000001"
			} else {
				precision, ok = safeGetString(args[2])
				if !ok {
					client.Emit("error", map[string]interface{}{
						"message": "Invalid precision format",
					})
					return
				}
			}
			room = getRoomName(subType, cast.ToString(poolID), "", precision)
		} else {
			room = getRoomName(subType, cast.ToString(poolID), "", "")
		}

		if room == "" {
			return
		}

		client.Leave(socket.Room(room))
		client.Emit("unsubscribed", map[string]interface{}{
			"type":    subType,
			"pool_id": poolID,
		})
	})

	client.On("unsubscribe_all", func(args ...interface{}) {
		rooms := client.Rooms()

		for _, room := range rooms.Keys() {
			if !strings.HasPrefix(string(room), "account:") {
				client.Leave(room)
			}
		}

		client.Emit("unsubscribed_all", map[string]interface{}{
			"status":  "success",
			"message": "Unsubscribed from all channels",
		})
	})

	client.On("unsubscribe_all_kline", func(args ...interface{}) {
		rooms := client.Rooms()
		unsubCount := 0

		for _, room := range rooms.Keys() {
			roomStr := string(room)
			if strings.HasPrefix(roomStr, string(SubTypeKline)+":") {
				client.Leave(room)
				unsubCount++
			}
		}

		client.Emit("unsubscribed_all_kline", map[string]interface{}{
			"status":  "success",
			"message": fmt.Sprintf("Unsubscribed from all kline data, total %d", unsubCount),
		})
	})

	client.On("unsubscribe_all_depth", func(args ...interface{}) {
		rooms := client.Rooms()
		unsubCount := 0

		for _, room := range rooms.Keys() {
			roomStr := string(room)
			if strings.HasPrefix(roomStr, string(SubTypeDepth)+":") {
				client.Leave(room)
				unsubCount++
			}
		}

		client.Emit("unsubscribed_all_depth", map[string]interface{}{
			"status":  "success",
			"message": fmt.Sprintf("Unsubscribed from all depth data, total %d", unsubCount),
		})
	})

	client.On("unsubscribe_all_pool_stats", func(args ...interface{}) {
		rooms := client.Rooms()

		for _, room := range rooms.Keys() {
			if room == "all_pool_stats" {
				client.Leave(room)
				break
			}
		}

		client.Emit("unsubscribed_all_pool_stats", map[string]interface{}{
			"status":  "success",
			"message": "Unsubscribed from all pool stats data",
		})
	})
}

// Push message to specific user
func (s *Server) PushToAccount(account string, event string, data interface{}) {
	if s.io == nil || account == "" || event == "" {
		return
	}
	log.Printf("push to account: %s-%s", account, event)
	accountRoom := socket.Room(fmt.Sprintf("account:%s", account))
	s.io.To(accountRoom).Emit(event, data)
}

// Broadcast message to subscribers
func (s *Server) Broadcast(sub Subscription, event string, data interface{}) {
	if s.io == nil || event == "" {
		return
	}
	room := socket.Room(getRoomName(string(sub.Type), cast.ToString(sub.PoolID), sub.Interval, sub.Precision))
	if room == "" {
		return
	}
	s.io.To(room).Emit(event, data)
}

// Close server
func (s *Server) Close() error {
	if s.consumer != nil {
		s.consumer.Stop()
	}
	if s.pusher != nil {
		s.pusher.Stop()
	}
	if s.io != nil {
		s.io.Close(nil)
	}
	return nil
}
