package points

import (
	"encoding/json"
	"exapp-go/config"
	"exapp-go/internal/db/ckhdb"
	"exapp-go/internal/db/db"
	"exapp-go/internal/entity"
	"exapp-go/pkg/nsqutil"
	"github.com/nsqio/go-nsq"
	"github.com/robfig/cron/v3"
	"log"
)

const (
	MsgTypeTradeUpdate = "trade_update"
)

type UserPointsService struct {
	ckhRepo  *ckhdb.ClickHouseRepo
	repo     *db.Repo
	consumer *nsqutil.Consumer
}

// Base NSQ message structure
type NSQMessage struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

func NewService() *UserPointsService {
	ckhRepo := ckhdb.New()
	repo := db.New()
	cfg := config.Conf()
	consumer := nsqutil.NewConsumer(cfg.Nsq.Lookupd, cfg.Nsq.LookupTTl)
	
	return &UserPointsService{
		repo:     repo,
		ckhRepo:  ckhRepo,
		consumer: consumer,
	}
}

func (s *UserPointsService) RunConsumer() error {
	
	err := s.consumer.Consume(MsgTypeTradeUpdate, "points", s.HandleMessage)
	if err != nil {
		
		log.Printf("Consume action sync failed: %v", err)
		return err
	}
	
	return nil
}

func (s *UserPointsService) CheckTrade() {

}

func (s *UserPointsService) HandleMessage(msg *nsq.Message) error {
	var nsqMsg NSQMessage
	if err := json.Unmarshal(msg.Body, &nsqMsg); err != nil {
		log.Printf("Failed to unmarshal NSQ message: %v,%v", err, string(msg.Body))
		return nil // Return nil to confirm message
	}
	
	// 非交易消息不处理
	if nsqMsg.Type != MsgTypeTradeUpdate {
		
		return nil
	}
	
	var tradeData entity.Trade
	if err := json.Unmarshal(nsqMsg.Data, &tradeData); err != nil {
		log.Printf("Failed to unmarshal trade data: %v", err)
		return nil
	}
	
	return nil
}

func Start() error {
	service := NewService()
	if err := service.RunConsumer(); err != nil {
		return err
	}
	
	c := cron.New(
		cron.WithSeconds(),
		cron.WithChain(cron.SkipIfStillRunning(cron.DefaultLogger)),
	)
	
	// 扫描交易
	if _, err := c.AddFunc("@every 3s", service.CheckTrade); err != nil {
		
		return err
	}
	
	c.Run()
	return nil
}
