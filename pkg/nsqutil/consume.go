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

type Worker struct {
	addr    string
	cfg     *nsq.Config
	channel string
	ncs     map[string]*nsq.Consumer
}

func NewWorker(channel string, lookupd string, lookupTTl time.Duration) *Worker {
	nCfg := nsq.NewConfig()
	nCfg.LookupdPollInterval = lookupTTl
	return &Worker{
		addr:    lookupd,
		cfg:     nCfg,
		channel: channel,
		ncs:     map[string]*nsq.Consumer{},
	}
}

func (w *Worker) Consume(topic string, handler nsq.HandlerFunc) error {
	if nc, ok := w.ncs[topic]; ok {
		nc.AddHandler(handler)
		return nil
	}

	nc, err := nsq.NewConsumer(topic, w.channel, w.cfg)
	if err != nil {
		return err
	}

	nc.SetLogger(log.New(io.Discard, "", log.LstdFlags), nsq.LogLevelWarning)
	nc.AddHandler(handler)

	if err := nc.ConnectToNSQLookupd(w.addr); err != nil {
		return err
	}

	w.ncs[topic] = nc
	return nil
}

func (w *Worker) StopConsume() {
	for _, nc := range w.ncs {
		nc.Stop()
	}
}
