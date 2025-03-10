package marketplace

import (
	"context"
	"exapp-go/config"
	"exapp-go/internal/db/ckhdb"
	"exapp-go/internal/db/db"
	"exapp-go/internal/entity"
	"fmt"
	"log"
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
	isBlacklisted := false
	for _, blacklistUID := range s.cfg.TradingCompetition.Blacklist {
		if uid == blacklistUID {
			isBlacklisted = true
			break
		}
	}

	records, err := s.repo.GetUserDayProfitRanking(ctx, dayTime, 10, s.cfg.TradingCompetition.Blacklist)
	if err != nil {
		return nil, err
	}

	var userRecord *db.UserDayProfitRecord
	var userRank int
	if uid != "" {
		userRecord, userRank, err = s.repo.GetUserDayProfitRankAndProfit(ctx, dayTime, uid, s.cfg.TradingCompetition.Blacklist)
		if err != nil {
			return nil, err
		}
	}

	uids := make([]string, 0, len(records))
	for _, record := range records {
		uids = append(uids, record.UID)
	}
	if uid != "" {
		uids = append(uids, uid)
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

	if uid != "" {
		if isBlacklisted {
			if user, exists := userMap[uid]; exists {
				result.Avatar = user.Avatar
			}
			result.UserProfit = "0"
			result.Rank = 0
			result.Point = 0
		}
		if userRecord != nil {
			result.UserProfit = userRecord.Profit.String()
			result.Rank = userRank
			if userRank > 0 && userRank <= len(s.cfg.TradingCompetition.DailyPoints) {
				result.Point = s.cfg.TradingCompetition.DailyPoints[userRank-1]
			}
		}
		if user, exists := userMap[uid]; exists {
			result.Avatar = user.Avatar
		}
	}

	return result, nil
}

func (s *TradeCompetitionService) GetAccumulatedProfitRanking(ctx context.Context, beginTime, endTime time.Time, uid string) (*entity.UserProfitRank, error) {
	isBlacklisted := false
	for _, blacklistUID := range s.cfg.TradingCompetition.Blacklist {
		if uid == blacklistUID {
			isBlacklisted = true
			break
		}
	}

	records, err := s.repo.GetUserAccumulatedProfitRanking(ctx, beginTime, endTime, 10, s.cfg.TradingCompetition.Blacklist)
	if err != nil {
		return nil, err
	}

	var userRecord *db.UserAccumulatedProfitRecord
	var userRank int
	if uid != "" {
		userRecord, userRank, err = s.repo.GetUserAccumulatedProfitRankAndProfit(ctx, beginTime, endTime, uid, s.cfg.TradingCompetition.Blacklist)
		if err != nil {
			return nil, err
		}
	}

	uids := make([]string, 0, len(records))
	for _, record := range records {
		uids = append(uids, record.UID)
	}
	if uid != "" {
		uids = append(uids, uid)
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
	if uid != "" {
		if isBlacklisted {
			if user, exists := userMap[uid]; exists {
				result.Avatar = user.Avatar
			}
			result.UserProfit = "0"
			result.Rank = 0
			result.Point = 0
		}
		if userRecord != nil {
			result.UserProfit = userRecord.Profit.String()
			result.Rank = userRank
			if userRank > 0 && userRank <= len(s.cfg.TradingCompetition.AccumulatedPoints) {
				result.Point = s.cfg.TradingCompetition.AccumulatedPoints[userRank-1]
			}
		}
		if user, exists := userMap[uid]; exists {
			result.Avatar = user.Avatar
		}
	}

	return result, nil
}

func (s *TradeCompetitionService) GetTotalTradeStats(ctx context.Context, uid string) (*entity.TotalTradeStats, error) {
	startTime := time.Now()
	defer func() {
		log.Printf("[TradeCompetition] GetTotalTradeStats completed in %v", time.Since(startTime))
	}()

	log.Printf("[TradeCompetition] Starting GetTotalTradeStats for uid: %s", uid)
	beginTime := s.cfg.TradingCompetition.BeginTime
	endTime := s.cfg.TradingCompetition.EndTime

	userPoints := 0
	if uid != "" {
		log.Printf("[TradeCompetition] Checking blacklist for uid: %s", uid)
		for _, blacklistUID := range s.cfg.TradingCompetition.Blacklist {
			if uid == blacklistUID {
				userPoints = 0
				log.Printf("[TradeCompetition] User %s found in blacklist", uid)
				break
			}
		}

		log.Printf("[TradeCompetition] Getting points for user: %s", uid)
		points, err := s.repo.GetUserTotalPoints(ctx, uid, beginTime, endTime)
		if err != nil {
			log.Printf("[TradeCompetition] Error getting user points: %v", err)
			return nil, fmt.Errorf("failed to get user points: %w", err)
		}
		userPoints = points
		log.Printf("[TradeCompetition] User points retrieved: %d", points)
	}

	log.Printf("[TradeCompetition] Getting total issued points")
	totalPoints := 0
	totalPoints, err := s.repo.GetIssuedPoints(ctx, beginTime, endTime)
	if err != nil {
		log.Printf("[TradeCompetition] Error getting total points: %v", err)
		return nil, fmt.Errorf("failed to get total points: %w", err)
	}
	log.Printf("[TradeCompetition] Total issued points retrieved: %d", totalPoints)

	log.Printf("[TradeCompetition] Getting faucet claim count")
	claimFaucetCount, err := s.repo.ClaimFaucetCount(ctx)
	if err != nil {
		log.Printf("[TradeCompetition] Error getting claim faucet count: %v", err)
		return nil, fmt.Errorf("failed to get claim faucet count: %w", err)
	}
	log.Printf("[TradeCompetition] Faucet claim count retrieved: %d", claimFaucetCount)

	totalPoints += int(claimFaucetCount) * s.cfg.TradingCompetition.FaucetPoint

	var userClaimFaucet bool
	if uid != "" {
		log.Printf("[TradeCompetition] Checking if user claimed faucet: %s", uid)
		claimed, err := s.repo.IsUserClaimFaucet(ctx, uid)
		if err != nil {
			log.Printf("[TradeCompetition] Error checking user faucet claim: %v", err)
			return nil, err
		}
		if claimed {
			userPoints += s.cfg.TradingCompetition.FaucetPoint
			log.Printf("[TradeCompetition] User has claimed faucet, added %d points", s.cfg.TradingCompetition.FaucetPoint)
		}
		userClaimFaucet = claimed
	}

	log.Printf("[TradeCompetition] Getting trade count and volume")
	_, tradeVolume, err := s.ckhRepo.GetTradeCountAndVolume(ctx)
	if err != nil {
		log.Printf("[TradeCompetition] Error getting trade count and volume: %v", err)
		return nil, err
	}
	log.Printf("[TradeCompetition] Trade volume retrieved: %v", tradeVolume)

	return &entity.TotalTradeStats{
		UserClaimedFaucet: userClaimFaucet,
		UserPoints:        userPoints,
		TotalPointsIssued: totalPoints,
		TotalTradeVolume:  tradeVolume,
		TotalTradeUser:    claimFaucetCount,
	}, nil
}
