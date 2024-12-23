package nsqutil

import (
	"exapp-go/pkg/utils"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/nsqio/go-nsq"
)

const retry = 3

type Client struct {
	nps       []*nsq.Producer
	tick      int
	tickMutex sync.RWMutex
}

func PublishClient(nsqdAddrs []string) *Client {
	nCfg := nsq.NewConfig()

	var nps []*nsq.Producer

	for _, addr := range nsqdAddrs {
		np, err := nsq.NewProducer(addr, nCfg)
		if err != nil {
			panic(any(err))
		}
		nps = append(nps, np)
	}

	return &Client{
		nps:  nps,
		tick: 0,
	}
}

func (c *Client) assignProducer() *nsq.Producer {
	c.tickMutex.Lock()
	defer c.tickMutex.Unlock()

	assigned := c.tick + 1
	if assigned >= len(c.nps) {
		assigned = 0
	}
	c.tick = assigned
	return c.nps[assigned]
}

func (c *Client) Publish(topic string, payload interface{}) error {
	data, err := utils.Marshal(payload)
	if err != nil {
		return err
	}
	dataStr := strings.Trim(string(data), `"`)
	err = c.publish(topic, []byte(dataStr), 3)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) publish(topic string, data []byte, retry int) error {
	if err := c.assignProducer().Publish(topic, data); err != nil {
		if retry > 0 {
			retry--
			return c.publish(topic, data, retry)
		}
		return err
	}

	return nil
}

func (c *Client) DeferredPublish(topic string, delay time.Duration, payload interface{}) error {
	data, err := utils.Marshal(payload)
	if err != nil {
		return err
	}

	err = c.deferredPublish(topic, delay, data, retry)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) deferredPublish(topic string, delay time.Duration, data []byte, retry int) error {
	if err := c.assignProducer().DeferredPublish(topic, delay, data); err != nil {
		if match, _ := regexp.MatchString("E_INVALID DPUB timeout \\d+ out of range 0-\\d+", err.Error()); match {
			return err
		}
		if retry > 0 {
			retry--
			return c.deferredPublish(topic, delay, data, retry)
		}
		return err
	}

	return nil
}
