package hyperion

import (
	"context"
	"fmt"
	"log"

	"github.com/gorilla/websocket"
)

type StreamClient struct {
	endpoint string
	conn     *websocket.Conn
}

func NewStreamClient(endpoint string) (*StreamClient, error) {

	conn, _, err := websocket.DefaultDialer.Dial(fmt.Sprintf("%s/stream", endpoint), nil)
	if err != nil {
		return nil, fmt.Errorf("dial websocket failed: %w", err)
	}

	return &StreamClient{
		endpoint: endpoint,
		conn:     conn,
	}, nil
}

type ActionStreamRequest struct {
	Contract  string          `json:"contract"`
	Action    string          `json:"action"`
	Account   string          `json:"account,omitempty"`
	StartFrom int64           `json:"start_from,omitempty"`
	ReadUntil int64           `json:"read_until,omitempty"`
	Filters   []RequestFilter `json:"filters,omitempty"`
}

type RequestFilter struct {
	Field string `json:"field"`
	Value string `json:"value"`
}

type ActionStreamResponse struct {
	Type    string   `json:"type"`
	Message []Action `json:"message"`
}

func (c *StreamClient) SubscribeAction(ctx context.Context, req ActionStreamRequest) (<-chan Action, error) {

	message := struct {
		Type    string              `json:"type"`
		Request ActionStreamRequest `json:"req"`
	}{
		Type:    "action_stream_request",
		Request: req,
	}

	if err := c.conn.WriteJSON(message); err != nil {
		return nil, fmt.Errorf("send subscribe request failed: %w", err)
	}

	actionCh := make(chan Action, 100)

	go func() {
		defer close(actionCh)
		defer c.conn.Close()

		for {
			select {
			case <-ctx.Done():
				return
			default:
				var resp ActionStreamResponse
				err := c.conn.ReadJSON(&resp)
				if err != nil {
					log.Printf("read message failed: %v", err)
					return
				}

				if resp.Type == "action_trace" {
					for _, action := range resp.Message {
						select {
						case actionCh <- action:
						default:
							log.Printf("action channel is full, dropping message")
						}
					}
				}
			}
		}
	}()

	return actionCh, nil
}

func (c *StreamClient) Close() error {
	return c.conn.Close()
}
