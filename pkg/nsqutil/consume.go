package nsqutil

import (
	"io"
	"log"
	"time"

	"github.com/nsqio/go-nsq"
)

func IsNsqMessageExpired(message *nsq.Message, expiredSecond float64) bool {
	return time.Since(time.Unix(0, message.Timestamp)).Seconds() > expiredSecond
}

type Consumer struct {
	addr string
	cfg  *nsq.Config
	ncs  map[string]map[string]*nsq.Consumer
}

func NewConsumer(lookupd string, lookupTTl time.Duration) *Consumer {
	nCfg := nsq.NewConfig()
	nCfg.LookupdPollInterval = lookupTTl
	return &Consumer{
		addr: lookupd,
		cfg:  nCfg,
		ncs:  make(map[string]map[string]*nsq.Consumer),
	}
}

func (w *Consumer) Consume(topic string, channel string, handler nsq.HandlerFunc) error {
	if _, ok := w.ncs[topic]; !ok {
		w.ncs[topic] = make(map[string]*nsq.Consumer)
	}

	if nc, ok := w.ncs[topic][channel]; ok {
		nc.AddHandler(handler)
		return nil
	}

	nc, err := nsq.NewConsumer(topic, channel, w.cfg)
	if err != nil {
		return err
	}

	nc.SetLogger(log.New(io.Discard, "", log.LstdFlags), nsq.LogLevelWarning)
	nc.AddHandler(handler)

	if err := nc.ConnectToNSQLookupd(w.addr); err != nil {
		return err
	}

	w.ncs[topic][channel] = nc
	return nil
}

func (w *Consumer) StopConsumeByTopic(topic string) {
	if nc, ok := w.ncs[topic]; ok {
		for _, nc := range nc {
			nc.Stop()
		}
		delete(w.ncs, topic)
	}
}




func (w *Consumer) Stop() {
	for _, nc := range w.ncs {
		for _, nc := range nc {
			nc.Stop()
		}
	}
	w.ncs = make(map[string]map[string]*nsq.Consumer)
}
