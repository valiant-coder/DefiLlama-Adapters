package points

import (
	"context"
	"encoding/json"
	"exapp-go/config"
	"exapp-go/internal/db/ckhdb"
	"exapp-go/internal/db/db"
	"exapp-go/internal/types"
	"exapp-go/pkg/log"
	"exapp-go/pkg/nsqutil"

	"github.com/nsqio/go-nsq"
	"github.com/robfig/cron/v3"
)

const (
	MsgTypeTradeUpdate = "trade_update"
)

type UserPointsService struct {
	ctx      context.Context
	repo     *db.Repo
	ckhRepo  *ckhdb.ClickHouseRepo
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
		ctx:      context.TODO(),
		repo:     repo,
		ckhRepo:  ckhRepo,
		consumer: consumer,
	}
}

func (s *UserPointsService) RunConsumer() error {

	if err := s.consumer.Consume(MsgTypeTradeUpdate, "points", s.HandleMessage); err != nil {

		log.Logger().Error("Consume action sync failed: %v", err)
		return err
	}

	return nil
}

func (s *UserPointsService) CheckTrade() {

}

func (s *UserPointsService) HandleMessage(msg *nsq.Message) error {
	var nsqMsg NSQMessage
	if err := json.Unmarshal(msg.Body, &nsqMsg); err != nil {
		log.Logger().Error("Failed to unmarshal NSQ message: %v,%v", err, string(msg.Body))
		return nil // Return nil to confirm message
	}

	// 非交易消息不处理
	if nsqMsg.Type != MsgTypeTradeUpdate {

		return nil
	}

	var trade ckhdb.Trade
	if err := json.Unmarshal(nsqMsg.Data, &trade); err != nil {
		log.Logger().Error("Failed to unmarshal trade data: %v", err)
		return nil
	}

	taker, err := s.repo.GetUIDByEOSAccountAndPermission(s.ctx, trade.Taker, trade.TakerPermission)
	if err != nil {

		log.Logger().Error(trade.Taker, "@", trade.TakerPermission, "Failed to get buyer uid: %v", err)
		return err
	}

	maker, err := s.repo.GetUIDByEOSAccountAndPermission(s.ctx, trade.Maker, trade.MakerPermission)
	if err != nil {

		log.Logger().Error(trade.Maker, "@", trade.MakerPermission, "Failed to get seller uid: %v", err)
		return err
	}

	// 结算邀请积分
	if err = s.SettlePoints(taker, maker, trade); err != nil {
		log.Logger().Error(trade.PoolID, taker, "Failed to settle points: %v", err)
		return err
	}

	// 结算返佣
	if err = s.Rebate(taker, maker, trade); err != nil {
		log.Logger().Error(trade.PoolID, taker, "Failed to rebate points: %v", err)
		return err
	}

	return nil
}

func (s *UserPointsService) SettlePoints(taker, maker string, trade ckhdb.Trade) (err error) {
	// 先计算积分
	conf, err := s.repo.GetUserPointsConf(s.ctx)
	if err != nil {

		log.Logger().Error("get user points conf error ->", err)
		return
	}

	// 读取交易对权重
	pair, err := s.repo.GetUserPointsPair(s.ctx, trade.Symbol)
	if err != nil {

		log.Logger().Error("get user points pair error ->", err)
		return
	}

	basePoints := conf.BaseTradePoints
	quantity := uint64(trade.QuoteQuantity.IntPart())
	takerPoints := basePoints * conf.TakerWeight * pair.Coefficient * quantity
	makerPoints := basePoints * conf.MakerWeight * pair.Coefficient * quantity

	// 更新用户积分
	if err = s.SettlerInvitePoints(taker, takerPoints, trade); err != nil {

		log.Logger().Error("settle points error ->", err)
		return
	}

	if err = s.SettlerInvitePoints(maker, makerPoints, trade); err != nil {

		log.Logger().Error("settle points error ->", err)
		return
	}

	return
}

func (s *UserPointsService) SettlerInvitePoints(uid string, points uint64, trade ckhdb.Trade) (err error) {

	err = s.repo.Transaction(s.ctx, func(repo *db.Repo) error {

		if e := repo.AddTradeUserPoints(s.ctx, uid, trade.TxID, points, trade.GlobalSequence, types.UserPointsTypeTrade); e != nil {

			log.Logger().Error(trade.TxID, "@", uid, "add trade user points error ->", e)
			return e
		}

		// 查询邀请信息
		invitation, _ := repo.GetUserInvitation(s.ctx, uid)
		if invitation == nil {

			return nil
		}

		// 查询邀请人
		if e := repo.AddTradeUserPoints(s.ctx, invitation.Inviter, trade.TxID, points, trade.GlobalSequence, types.UserPointsTypeInvitation); e != nil {

			log.Logger().Error(trade.TxID, "@", invitation.Inviter, "add trade user points error ->", e)
			return e
		}

		return nil
	})
	return
}

func (s *UserPointsService) Rebate(taker, maker string, trade ckhdb.Trade) (err error) {

	return
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
