package marketplace

import (
	"context"
	"exapp-go/config"
	"exapp-go/data"
	"exapp-go/internal/db/ckhdb"
	"exapp-go/internal/db/db"
	"exapp-go/pkg/log"
	"exapp-go/pkg/nsqutil"
)

type UserPointsService struct {
	repo    *db.Repo
	ckhRepo *ckhdb.ClickHouseRepo
	nsqPub  *nsqutil.Publisher
}

func NewUserPointsService() *UserPointsService {
	nsqConf := config.Conf().Nsq
	return &UserPointsService{
		repo:    db.New(),
		ckhRepo: ckhdb.New(),
		nsqPub:  nsqutil.NewPublisher(nsqConf.Nsqds),
	}
}

func (s *UserPointsService) GetUserPoints(ctx context.Context, uid string) (*db.UserPoints, error) {
	userPoints, err := s.repo.GetUserPoints(ctx, uid)
	if err != nil {

		log.Logger().Error(uid, "查询用户积分信息出错 -> ", err.Error())
		return nil, err
	}

	return userPoints, nil
}

func (s *UserPointsService) GetUserPointsRecords(ctx context.Context, params data.UPRecordListParam) (*db.UPRecordListRes, error) {
	result, err := s.repo.ListUPRecords(ctx, params)
	if err != nil {

		log.Logger().Error("查询用户积分记录出错 -> ", err)
		return nil, err
	}

	return &result, nil
}

func (s *UserPointsService) GetUserPointsConf(ctx context.Context) (*db.UserPointsConf, error) {
	userPointsConf, err := s.repo.GetUserPointsConf(ctx)
	if err != nil {

		log.Logger().Error("获取积分配置信息出错 -> ", err)
		return nil, err
	}

	return userPointsConf, nil
}

func (s *UserPointsService) UpdateUserPointsConf(ctx context.Context, params *data.UserPointsConfParam) error {

	userPointsConf, _ := s.repo.GetUserPointsConf(ctx)

	if userPointsConf == nil {

		userPointsConf = &db.UserPointsConf{}
	}

	if params.BaseTradePoints > 0 {
		userPointsConf.BaseTradePoints = params.BaseTradePoints
	}

	if params.MakerWeight > 0 {
		userPointsConf.MakerWeight = params.MakerWeight
	}

	if params.TakerWeight > 0 {
		userPointsConf.TakerWeight = params.TakerWeight
	}

	if params.FirstTradeRate > 0 {
		userPointsConf.FirstTradeRate = params.FirstTradeRate
	}

	if params.MaxPerTradePoints > 0 {
		userPointsConf.MaxPerTradePoints = params.MaxPerTradePoints
	}

	if params.InvitePercent > 0 {
		userPointsConf.InvitePercent = params.InvitePercent
	}

	if params.InviteRebatePercent > 0 {
		userPointsConf.InviteRebatePercent = params.InviteRebatePercent
	}

	if params.MaxPerInvitePoints > 0 {
		userPointsConf.MaxPerInvitePoints = params.MaxPerInvitePoints
	}

	if params.UpgradeInviterCount > 0 {
		userPointsConf.UpgradeInviterCount = params.UpgradeInviterCount
	}

	if params.UpgradeInvitePercent > 0 {
		userPointsConf.UpgradeInvitePercent = params.UpgradeInvitePercent
	}

	if params.OrderMinPendingTime > 0 {
		userPointsConf.OrderMinPendingTime = params.OrderMinPendingTime
	}

	if params.OrderMinValue > 0 {
		userPointsConf.OrderMinValue = params.OrderMinValue
	}

	if params.MaxInviteLinkCount > 0 {
		userPointsConf.MaxInviteLinkCount = params.MaxInviteLinkCount
	}

	var err error
	if userPointsConf.ID == 0 {

		err = s.repo.Insert(ctx, userPointsConf)
	} else {

		err = s.repo.Update(ctx, userPointsConf)
	}

	if err != nil {

		log.Logger().Error("更新积分配置信息出错 -> ", err)
		return err
	}

	return nil
}
