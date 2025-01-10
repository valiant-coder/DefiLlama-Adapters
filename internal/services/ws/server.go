package ws

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/spf13/cast"
	"github.com/zishang520/engine.io/types"
	"github.com/zishang520/socket.io/v2/socket"
)

type Server struct {
	io     *socket.Server
	pusher *Pusher
}

// Create new WebSocket server
func NewServer(ctx context.Context) *Server {
	c := socket.DefaultServerOptions()
	c.SetServeClient(true)
	c.SetPingInterval(10 * time.Second)
	c.SetPingTimeout(5 * time.Second)
	c.SetMaxHttpBufferSize(1000000)
	c.SetConnectTimeout(5 * time.Second)
	c.SetCors(&types.Cors{
		Origin:      "*",
		Credentials: true,
	})

	io := socket.NewServer(nil, c)

	server := &Server{
		io: io,
	}
	server.pusher = NewPusher(ctx, server)
	// Set connection handler
	io.On("connection", server.handleConnection)

	return server
}

// Get HTTP handler
func (s *Server) Handler() http.Handler {
	return s.io.ServeHandler(nil)
}

// Generate room name
func getRoomName(subType, poolID, interval string) string {
	if interval != "" {
		return fmt.Sprintf("%s:%s:%s", subType, poolID, interval)
	}
	return fmt.Sprintf("%s:%s", subType, poolID)
}

// Handle new connection
func (s *Server) handleConnection(args ...interface{}) {
	client := args[0].(*socket.Socket)

	// User authentication
	client.On("authenticate", func(args ...interface{}) {
		if len(args) < 1 {
			return
		}
		account := args[0].(string)
		s.authenticateUser(client, account)
	})

	// Subscribe to kline data
	client.On("subscribe_kline", func(args ...interface{}) {
		if len(args) < 2 {
			return
		}
		poolID := uint64(args[0].(float64))
		interval := args[1].(string)
		room := socket.Room(getRoomName(string(SubTypeKline), cast.ToString(poolID), interval))
		client.Join(room)
		client.Emit("subscribed", map[string]interface{}{
			"type":     SubTypeKline,
			"poolID":   poolID,
			"interval": interval,
		})
	})

	// Subscribe to depth data
	client.On("subscribe_depth", func(args ...interface{}) {
		if len(args) < 1 {
			return
		}
		poolID := uint64(args[0].(float64))
		room := socket.Room(getRoomName(string(SubTypeDepth), cast.ToString(poolID), ""))
		client.Join(room)
		client.Emit("subscribed", map[string]interface{}{
			"type":   SubTypeDepth,
			"poolID": poolID,
		})
	})

	// Subscribe to trade data
	client.On("subscribe_trades", func(args ...interface{}) {
		if len(args) < 1 {
			return
		}
		poolID := uint64(args[0].(float64))

		room := socket.Room(getRoomName(string(SubTypeTrades), cast.ToString(poolID), ""))
		client.Join(room)
		client.Emit("subscribed", map[string]interface{}{
			"type":   SubTypeTrades,
			"poolID": poolID,
		})
	})

	// Unsubscribe
	client.On("unsubscribe", func(args ...interface{}) {
		if len(args) < 2 {
			return
		}
		poolID := uint64(args[0].(float64))
		subType := args[1].(string)
		var room string
		if subType == string(SubTypeKline) {
			if len(args) < 3 {
				return
			}
			interval := args[2].(string)
			room = getRoomName(subType, cast.ToString(poolID), interval)
		} else {
			room = getRoomName(subType, cast.ToString(poolID), "")
		}
		client.Leave(socket.Room(room))
		client.Emit("unsubscribed", map[string]interface{}{
			"type":   subType,
			"poolID": poolID,
		})
	})

}

// Authenticate user
func (s *Server) authenticateUser(client *socket.Socket, account string) {

	// Add user to dedicated room
	accountRoom := socket.Room(fmt.Sprintf("user:%s", account))
	client.Join(accountRoom)

	// Send authentication success message
	client.Emit("authenticated", map[string]interface{}{
		"status": "success",
		"userId": account,
	})
}

// Push message to specific user
func (s *Server) PushToAccount(account string, event string, data interface{}) {
	accountRoom := socket.Room(fmt.Sprintf("user:%s", account))
	s.io.To(accountRoom).Emit(event, data)
}

// Broadcast message to subscribers
func (s *Server) Broadcast(sub Subscription, event string, data interface{}) {
	room := socket.Room(getRoomName(string(sub.Type), cast.ToString(sub.PoolID), sub.Interval))
	s.io.To(room).Emit(event, data)
}

// Close server
func (s *Server) Close() error {
	s.pusher.Stop()
	s.io.Close(nil)
	return nil
}
