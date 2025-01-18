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

type Publisher struct {
	nps       []*nsq.Producer
	tick      int
	tickMutex sync.RWMutex
}

func NewPublisher(nsqdAddrs []string) *Publisher {
	nCfg := nsq.NewConfig()

	var nps []*nsq.Producer

	for _, addr := range nsqdAddrs {
		np, err := nsq.NewProducer(addr, nCfg)
		if err != nil {
			panic(any(err))
		}
		nps = append(nps, np)
	}

	return &Publisher{
		nps:  nps,
		tick: 0,
	}
}

func (c *Publisher) assignProducer() *nsq.Producer {
	c.tickMutex.Lock()
	defer c.tickMutex.Unlock()

	assigned := c.tick + 1
	if assigned >= len(c.nps) {
		assigned = 0
	}
	c.tick = assigned
	return c.nps[assigned]
}

func (c *Publisher) Publish(topic string, payload interface{}) error {
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

func (c *Publisher) publish(topic string, data []byte, retry int) error {
	if err := c.assignProducer().Publish(topic, data); err != nil {
		if retry > 0 {
			retry--
			return c.publish(topic, data, retry)
		}
		return err
	}

	return nil
}

func (c *Publisher) DeferredPublish(topic string, delay time.Duration, payload interface{}) error {
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

func (c *Publisher) deferredPublish(topic string, delay time.Duration, data []byte, retry int) error {
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

func (c *Publisher) Stop() {
	for _, np := range c.nps {
		np.Stop()
	}
}
