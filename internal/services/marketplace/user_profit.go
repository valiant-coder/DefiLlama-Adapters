package marketplace

import (
	"context"
	"exapp-go/internal/db/db"
	"exapp-go/internal/entity"
	"time"
)

type UserProfitService struct {
	repo *db.Repo
}

func NewUserProfitService() *UserProfitService {
	return &UserProfitService{
		repo: db.New(),
	}
}

func (s *UserProfitService) GetDayProfitRanking(ctx context.Context, dayTime time.Time, uid string) (*entity.UserProfitRank, error) {
	records, err := s.repo.GetUserDayProfitRanking(ctx, dayTime, 20)
	if err != nil {
		return nil, err
	}

	userRecord, userRank, err := s.repo.GetUserDayProfitRankAndProfit(ctx, dayTime, uid)
	if err != nil {
		return nil, err
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

	for _, record := range records {
		user, exists := userMap[record.UID]
		avatar := ""
		if exists {
			avatar = user.Avatar
		}
		result.Items = append(result.Items, entity.UserProfit{
			UID:    record.UID,
			Avatar: avatar,
			Profit: record.Profit.String(),
		})
	}

	if userRecord != nil {
		result.UserProfit = userRecord.Profit.String()
		result.Rank = userRank
		if user, exists := userMap[userRecord.UID]; exists {
			result.Avatar = user.Avatar
		}
	}

	return result, nil
}

func (s *UserProfitService) GetAccumulatedProfitRanking(ctx context.Context, beginTime, endTime time.Time, uid string) (*entity.UserProfitRank, error) {
	records, err := s.repo.GetUserAccumulatedProfitRanking(ctx, beginTime, endTime, 20)
	if err != nil {
		return nil, err
	}

	userRecord, userRank, err := s.repo.GetUserAccumulatedProfitRankAndProfit(ctx, beginTime, endTime, uid)
	if err != nil {
		return nil, err
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

	for _, record := range records {
		user, exists := userMap[record.UID]
		avatar := ""
		if exists {
			avatar = user.Avatar
		}
		result.Items = append(result.Items, entity.UserProfit{
			UID:    record.UID,
			Avatar: avatar,
			Profit: record.Profit.String(),
		})
	}

	if userRecord != nil {
		result.UserProfit = userRecord.Profit.String()
		result.Rank = userRank
		if user, exists := userMap[userRecord.UID]; exists {
			result.Avatar = user.Avatar
		}
	}

	return result, nil
}
