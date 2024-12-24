package hyperion

import (
	"fmt"
	"log"

	socketio_client "github.com/ahollic/go-socket.io-client"
)

type StreamClient struct {
	endpoint string
	client   *socketio_client.Client
}

func NewStreamClient(endpoint string) (*StreamClient, error) {
	opts := &socketio_client.Options{
		Transport: "websocket",
		Path:      "/stream",
		Query:     make(map[string]string),
	}

	client, err := socketio_client.NewClient(endpoint, opts)
	if err != nil {
		return nil, fmt.Errorf("create socket.io client failed: %w", err)
	}
	client.On("error", func() {
		log.Printf("socket.io connection error")
	})

	client.On("connect", func() {
		log.Printf("socket.io connected successfully")
	})

	
	return &StreamClient{
		endpoint: endpoint,
		client:   client,
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

func (c *StreamClient) SubscribeAction(req ActionStreamRequest) (<-chan Action, error) {
	actionCh := make(chan Action, 100)


	if err := c.client.Emit("action_stream_request", req); err != nil {
		return nil, fmt.Errorf("send subscribe request to hyperion failed: %w", err)
	}

	c.client.On("message", func(response ActionStreamResponse) {
		if response.Type == "action_trace" {
			for _, action := range response.Message {
				select {
				case actionCh <- action:
				default:
					log.Printf("action channel is full, dropping message")
				}
			}
		}
	})

	return actionCh, nil
}

func (c *StreamClient) Close() error {
	return nil
}
