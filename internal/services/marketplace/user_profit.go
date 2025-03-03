package marketplace

import (
	"context"
	"exapp-go/config"
	"exapp-go/internal/db/db"
	"exapp-go/internal/entity"
	"time"
)

type UserProfitService struct {
	repo *db.Repo
	cfg  *config.Config
}

func NewUserProfitService() *UserProfitService {
	return &UserProfitService{
		repo: db.New(),
		cfg:  config.Conf(),
	}
}

func (s *UserProfitService) GetDayProfitRanking(ctx context.Context, dayTime time.Time, uid string) (*entity.UserProfitRank, error) {
	records, err := s.repo.GetUserDayProfitRanking(ctx, dayTime, 20)
	if err != nil {
		return nil, err
	}

	var userRecord *db.UserDayProfitRecord
	var userRank int
	if uid != "" {
		userRecord, userRank, err = s.repo.GetUserDayProfitRankAndProfit(ctx, dayTime, uid)
		if err != nil {
			return nil, err
		}
	}

	uids := make([]string, 0, len(records))
	for _, record := range records {
		uids = append(uids, record.UID)
	}
	if userRecord != nil {
		uids = append(uids, userRecord.UID)
	}

	userMap, err := s.repo.GetUsersByUIDs(ctx, uids)
	if err != nil {
		return nil, err
	}

	result := &entity.UserProfitRank{
		Items: make([]entity.UserProfit, 0, len(records)),
	}

	for i, record := range records {
		user, exists := userMap[record.UID]
		avatar := ""
		if exists {
			avatar = user.Avatar
		}
		point := 0
		if i < len(s.cfg.TradingCompetition.DailyPoints) {
			point = s.cfg.TradingCompetition.DailyPoints[i]
		}
		result.Items = append(result.Items, entity.UserProfit{
			UID:    record.UID,
			Avatar: avatar,
			Profit: record.Profit.String(),
			Point:  point,
		})
	}

	if userRecord != nil {
		result.UserProfit = userRecord.Profit.String()
		result.Rank = userRank
		if user, exists := userMap[userRecord.UID]; exists {
			result.Avatar = user.Avatar
		}
		if userRank > 0 && userRank <= len(s.cfg.TradingCompetition.DailyPoints) {
			result.Point = s.cfg.TradingCompetition.DailyPoints[userRank-1]
		}
	}

	return result, nil
}

func (s *UserProfitService) GetAccumulatedProfitRanking(ctx context.Context, beginTime, endTime time.Time, uid string) (*entity.UserProfitRank, error) {
	records, err := s.repo.GetUserAccumulatedProfitRanking(ctx, beginTime, endTime, 20)
	if err != nil {
		return nil, err
	}

	var userRecord *db.UserAccumulatedProfitRecord
	var userRank int
	if uid != "" {
		userRecord, userRank, err = s.repo.GetUserAccumulatedProfitRankAndProfit(ctx, beginTime, endTime, uid)
		if err != nil {
			return nil, err
		}
	}

	uids := make([]string, 0, len(records))
	for _, record := range records {
		uids = append(uids, record.UID)
	}
	if userRecord != nil {
		uids = append(uids, userRecord.UID)
	}

	userMap, err := s.repo.GetUsersByUIDs(ctx, uids)
	if err != nil {
		return nil, err
	}

	result := &entity.UserProfitRank{
		Items: make([]entity.UserProfit, 0, len(records)),
	}

	for i, record := range records {
		user, exists := userMap[record.UID]
		avatar := ""
		if exists {
			avatar = user.Avatar
		}
		point := 0
		if i < len(s.cfg.TradingCompetition.AccumulatedPoints) {
			point = s.cfg.TradingCompetition.AccumulatedPoints[i]
		}
		result.Items = append(result.Items, entity.UserProfit{
			UID:    record.UID,
			Avatar: avatar,
			Profit: record.Profit.String(),
			Point:  point,
		})
	}

	if userRecord != nil {
		result.UserProfit = userRecord.Profit.String()
		result.Rank = userRank
		if user, exists := userMap[userRecord.UID]; exists {
			result.Avatar = user.Avatar
		}
		if userRank > 0 && userRank <= len(s.cfg.TradingCompetition.AccumulatedPoints) {
			result.Point = s.cfg.TradingCompetition.AccumulatedPoints[userRank-1]
		}
	}

	return result, nil
}
