package hyperion

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	socketio_client "github.com/LiterMC/socket.io"
	engine "github.com/LiterMC/socket.io/engine.io"
)

type StreamClient struct {
	endpoint string
	client   *socketio_client.Socket
}

func NewStreamClient(endpoint string) (*StreamClient, error) {
	engio, err := engine.NewSocket(
		engine.Options{
			Host: endpoint,
			Path: "/stream/",
		},
	)
	if err != nil {
		return nil, fmt.Errorf("create engine.io socket failed: %w", err)
	}

	client := socketio_client.NewSocket(
		engio,
	)
	client.OnConnect(func(s *socketio_client.Socket, namespace string) {
		log.Printf("socket.io connected successfully")
	})

	client.OnDisconnect(func(s *socketio_client.Socket, namespace string) {
		log.Printf("socket.io disconnected")
	})

	log.Printf("Dialing %s", engio.URL().String())
	if err := engio.Dial(context.Background()); err != nil {
		return nil, fmt.Errorf("dial error: %w", err)
	}
	log.Printf("Connecting to socket.io namespace")
	if err := client.Connect(""); err != nil {
		return nil, fmt.Errorf("open namespace error: %w", err)
	}

	return &StreamClient{
		endpoint: endpoint,
		client:   client,
	}, nil
}

type ActionStreamRequest struct {
	Contract  string          `json:"contract"`
	Action    string          `json:"action"`
	Account   string          `json:"account"`
	StartFrom int64           `json:"start_from"`
	ReadUntil int64           `json:"read_until"`
	Filters   []RequestFilter `json:"filters"`
}

type RequestFilter struct {
	Field string `json:"field"`
	Value string `json:"value"`
}

type ActionStreamResponse struct {
	Type    string   `json:"type"`
	Mode    string   `json:"mode"`
	Message []Action `json:"message"`
}

func (c *StreamClient) SubscribeAction(reqs []ActionStreamRequest) (<-chan Action, error) {
	actionCh := make(chan Action, 100)

	for _, req := range reqs {
		_, err := c.client.EmitWithAck("action_stream_request", req)
		if err != nil {
			return nil, fmt.Errorf("send subscribe request to hyperion failed: %w, req: %+v", err, req)
		}
	}

	c.client.OnMessage(func(event string, args []any) {
		if event != "message" {
			return
		}

		message := args[0].(map[string]any)
		messageType, ok := message["type"].(string)
		if !ok {
			return
		}
		if messageType != "action_trace" {
			return
		}
		mode, ok := message["mode"].(string)
		if !ok {
			return
		}
		if mode != "live" {
			return
		}

		var action Action
		if err := json.Unmarshal([]byte(message["message"].(string)), &action); err != nil {
			log.Printf("unmarshal response failed: %v", err)
			return
		}
		log.Printf("action: %+v", action)

		select {
		case actionCh <- action:
		default:
			log.Printf("action channel is full, dropping message")
		}

	})

	return actionCh, nil
}




type DeltaStreamRequest struct {
	Code string `json:"code"`
	Table string `json:"table"`
	Scope string `json:"scope"`
	Payer string `json:"payer"`
	StartFrom int64 `json:"start_from"`
	ReadUntil int64 `json:"read_until"`
}


type DeltaStreamResponse struct {
	Type string `json:"type"`
	Mode string `json:"mode"`
	Message []Delta `json:"message"`
}

func (c *StreamClient) SubscribeDeltas(reqs []DeltaStreamRequest) (<-chan Delta, error) {
	deltaCh := make(chan Delta, 100)

	for _, req := range reqs {
		_, err := c.client.EmitWithAck("delta_stream_request", req)
		if err != nil {
			return nil, fmt.Errorf("send subscribe request to hyperion failed: %w, req: %+v", err, req)
		}
	}

	c.client.OnMessage(func(event string, args []any) {
		if event != "message" {
			return
		}

		message := args[0].(map[string]any)
		messageType, ok := message["type"].(string)
		if !ok {
			return
		}
		if messageType != "delta_trace" {
			return
		}

		var delta Delta
		if err := json.Unmarshal([]byte(message["message"].(string)), &delta); err != nil {
			log.Printf("unmarshal response failed: %v", err)
			return
		}

		select {
		case deltaCh <- delta:
		default:
			log.Printf("delta channel is full, dropping message")
		}
	})

	return deltaCh, nil
}




func (c *StreamClient) Close() error {
	return c.client.Close()
}
