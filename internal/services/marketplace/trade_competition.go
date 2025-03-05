package marketplace

import (
	"context"
	"exapp-go/config"
	"exapp-go/internal/db/ckhdb"
	"exapp-go/internal/db/db"
	"exapp-go/internal/entity"
	"fmt"
	"time"
)

type TradeCompetitionService struct {
	repo    *db.Repo
	ckhRepo *ckhdb.ClickHouseRepo
	cfg     *config.Config
}

func NewTradeCompetitionService() *TradeCompetitionService {
	return &TradeCompetitionService{
		repo:    db.New(),
		ckhRepo: ckhdb.New(),
		cfg:     config.Conf(),
	}
}

func (s *TradeCompetitionService) GetDayProfitRanking(ctx context.Context, dayTime time.Time, uid string) (*entity.UserProfitRank, error) {
	records, err := s.repo.GetUserDayProfitRanking(ctx, dayTime, 10)
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
			Rank:   i + 1,
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

func (s *TradeCompetitionService) GetAccumulatedProfitRanking(ctx context.Context, beginTime, endTime time.Time, uid string) (*entity.UserProfitRank, error) {
	records, err := s.repo.GetUserAccumulatedProfitRanking(ctx, beginTime, endTime, 10)
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
			Rank:   i + 1,
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

func (s *TradeCompetitionService) GetTotalTradeStats(ctx context.Context, uid string) (*entity.TotalTradeStats, error) {
	beginTime := s.cfg.TradingCompetition.BeginTime
	endTime := s.cfg.TradingCompetition.EndTime

	userPoints := 0
	if uid != "" {
		points, err := s.repo.GetUserTotalPoints(ctx, uid, beginTime, endTime)
		if err != nil {
			return nil, fmt.Errorf("failed to get user points: %w", err)
		}
		userPoints = points
	}

	totalPoints := 0
	totalPoints, err := s.repo.GetIssuedPoints(ctx, beginTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to get total points: %w", err)
	}

	claimFaucetCount, err := s.repo.ClaimFaucetCount(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get claim faucet count: %w", err)
	}

	totalPoints += int(claimFaucetCount) * s.cfg.TradingCompetition.FaucetPoint

	var userClaimFaucet bool
	if uid != "" {
		claimed, err := s.repo.IsUserClaimFaucet(ctx, uid)
		if err != nil {
			return nil, err
		}
		if claimed {
			userPoints += s.cfg.TradingCompetition.FaucetPoint
		}
		userClaimFaucet = claimed
	}

	_, tradeVolume, err := s.ckhRepo.GetTradeCountAndVolume(ctx)
	if err != nil {
		return nil, err
	}
	totalUserCount, err := s.repo.GetTotalUserCount(ctx)
	if err != nil {
		return nil, err
	}

	return &entity.TotalTradeStats{
		UserClaimedFaucet: userClaimFaucet,
		UserPoints:        userPoints,
		TotalPointsIssued: totalPoints,
		TotalTradeVolume:  tradeVolume,
		TotalTradeUser:    totalUserCount,
	}, nil
}
