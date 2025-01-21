package handler

import (
	"context"
	"encoding/json"
	"exapp-go/config"
	"exapp-go/internal/db/ckhdb"
	"exapp-go/internal/db/db"
	"exapp-go/pkg/hyperion"
	"exapp-go/pkg/nsqutil"
	"fmt"
	"hash/fnv"
	"log"
	"sort"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/nsqio/go-nsq"
	"github.com/redis/go-redis/v9"
)

const (
	TopicActionSync = "cdex_action_sync"

	// Redis keys
	RedisKeyHandlerInstances = "cdex:handler:instances" // Hash table stores all handler instances
	RedisKeyHandlerLock      = "cdex:handler:lock"      // Distributed lock key
	HandlerTTL               = 4 * time.Second          // Handler heartbeat timeout
)

type Service struct {
	ckhRepo     *ckhdb.ClickHouseRepo
	repo        *db.Repo
	consumer    *nsqutil.Consumer
	poolCache   map[uint64]*db.Pool
	eosCfg      config.EosConfig
	cdexCfg     config.CdexConfig
	exappCfg    config.ExappConfig
	exsatCfg    config.ExsatConfig
	publisher   *NSQPublisher
	redisCli    redis.Cmdable
	instanceID  string
	curInstance int
	klineCache  map[uint64]map[ckhdb.KlineInterval]*ckhdb.Kline // Cache latest kline data for each trading pair's intervals
	handlers    map[string]func(hyperion.Action) error
}

func NewService() (*Service, error) {
	ckhRepo := ckhdb.New()
	repo := db.New()
	cfg := config.Conf()

	publisher, err := NewNSQPublisher(cfg.Nsq.Nsqds)
	if err != nil {
		return nil, err
	}

	redisCli := repo.Redis()
	// Generate unique instance ID
	instanceID := uuid.New().String()
	consumer := nsqutil.NewConsumer(cfg.Nsq.Lookupd, cfg.Nsq.LookupTTl)

	s := &Service{
		ckhRepo:    ckhRepo,
		repo:       repo,
		poolCache:  make(map[uint64]*db.Pool),
		eosCfg:     cfg.Eos,
		cdexCfg:    cfg.Eos.CdexConfig,
		exappCfg:   cfg.Eos.Exapp,
		exsatCfg:   cfg.Eos.Exsat,
		publisher:  publisher,
		redisCli:   redisCli,
		instanceID: instanceID,
		consumer:   consumer,
		klineCache: make(map[uint64]map[ckhdb.KlineInterval]*ckhdb.Kline),
		handlers:   make(map[string]func(hyperion.Action) error),
	}

	// Register all handlers
	s.registerHandlers()

	return s, nil
}

func (s *Service) Start(ctx context.Context) error {

	s.redisCli.HSet(ctx, RedisKeyHandlerInstances, s.instanceID, time.Now().Unix())
	// Start heartbeat goroutine
	go s.heartbeat(ctx)

	// Initialize kline cache
	if err := s.initKlineCache(ctx); err != nil {
		log.Printf("init kline cache failed: %v", err)
	}

	err := s.consumer.Consume(TopicActionSync, fmt.Sprintf("%s#ephemeral", uuid.New().String()), s.HandleMessage)
	if err != nil {
		log.Printf("Consume action sync failed: %v", err)
		return err
	}

	<-ctx.Done()
	return nil
}

func (s *Service) Stop(ctx context.Context) error {
	// Remove instance info from Redis
	s.redisCli.HDel(ctx, RedisKeyHandlerInstances, s.instanceID)

	s.consumer.Stop()
	if s.publisher != nil {
		s.publisher.Close()
	}
	return nil
}

// heartbeat periodically updates instance info in Redis
func (s *Service) heartbeat(ctx context.Context) {
	ticker := time.NewTicker(HandlerTTL / 2)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Update instance heartbeat time
			err := s.redisCli.HSet(ctx, RedisKeyHandlerInstances, s.instanceID, time.Now().Unix()).Err()
			if err != nil {
				log.Printf("Update handler heartbeat failed: %v", err)
			}

			// Clean expired instances
			s.cleanExpiredInstances(ctx)
		}
	}
}

// cleanExpiredInstances cleans expired instances
func (s *Service) cleanExpiredInstances(ctx context.Context) {
	// Get distributed lock
	lock := s.redisCli.SetNX(ctx, RedisKeyHandlerLock, s.instanceID, HandlerTTL)
	if !lock.Val() {
		return
	}
	defer s.redisCli.Del(ctx, RedisKeyHandlerLock)

	// Get all instances
	instances, err := s.redisCli.HGetAll(ctx, RedisKeyHandlerInstances).Result()
	if err != nil {
		log.Printf("Get handler instances failed: %v", err)
		return
	}

	now := time.Now().Unix()
	for id, lastHeartbeat := range instances {
		heartbeat, _ := strconv.ParseInt(lastHeartbeat, 10, 64)
		if now-heartbeat > int64(HandlerTTL.Seconds()) {
			s.redisCli.HDel(ctx, RedisKeyHandlerInstances, id)
		}
	}
}

// getInstanceInfo gets current instance ID and total instances count
func (s *Service) getInstanceInfo(ctx context.Context) (int, int) {
	// Get all active instances
	instances, err := s.redisCli.HGetAll(ctx, RedisKeyHandlerInstances).Result()
	if err != nil {
		log.Printf("Get handler instances failed: %v", err)
		return 0, 1
	}

	// Sort instance IDs to ensure consistent order across instances
	instanceIDs := make([]string, 0, len(instances))
	for id := range instances {
		instanceIDs = append(instanceIDs, id)
	}
	sort.Strings(instanceIDs)

	// Find current instance index
	currentInstance := 0
	for i, id := range instanceIDs {
		if id == s.instanceID {
			currentInstance = i
			break
		}
	}
	if s.curInstance != currentInstance {
		s.curInstance = currentInstance
		s.initKlineCache(ctx)
	}

	return currentInstance, len(instances)
}

func (s *Service) registerHandlers() {
	// Use account:action as key
	s.handlers[fmt.Sprintf("%s:%s", s.cdexCfg.EventContract, s.eosCfg.Events.EmitPlaced)] = s.handleCreateOrder
	s.handlers[fmt.Sprintf("%s:%s", s.cdexCfg.EventContract, s.eosCfg.Events.EmitCanceled)] = s.handleCancelOrder
	s.handlers[fmt.Sprintf("%s:%s", s.cdexCfg.EventContract, s.eosCfg.Events.EmitFilled)] = s.handleMatchOrder
	s.handlers[fmt.Sprintf("%s:%s", s.cdexCfg.PoolContract, s.eosCfg.Events.Create)] = s.handleCreatePool
	s.handlers[fmt.Sprintf("%s:%s", s.exsatCfg.BridgeContract, s.eosCfg.Events.DepositLog)] = s.handleBridgeDeposit
	s.handlers[fmt.Sprintf("%s:%s", s.exappCfg.AssetContract, s.eosCfg.Events.LogNewAcc)] = s.handleNewAccount
	s.handlers[fmt.Sprintf("%s:%s", s.exappCfg.AssetContract, s.eosCfg.Events.LogWithdraw)] = s.handleWithdraw
	s.handlers[fmt.Sprintf("%s:%s", s.exsatCfg.BridgeContract, s.eosCfg.Events.WithdrawLog)] = s.updateWithdraw
}

func (s *Service) HandleMessage(msg *nsq.Message) error {
	log.Println("get new action")
	var action hyperion.Action
	if err := json.Unmarshal(msg.Body, &action); err != nil {
		log.Printf("Unmarshal action failed: %v", err)
		return nil
	}

	// Check if should process this message
	if action.Act.Account != s.cdexCfg.EventContract &&
		action.Act.Account != s.cdexCfg.PoolContract &&
		action.Act.Account != s.exappCfg.AssetContract &&
		action.Act.Account != s.exsatCfg.BridgeContract {
		return nil
	}

	// Get handler key
	handlerKey := fmt.Sprintf("%s:%s", action.Act.Account, action.Act.Name)

	// Find corresponding handler
	handler, exists := s.handlers[handlerKey]
	if !exists {
		log.Printf("Unknown action: %s from account: %s", action.Act.Name, action.Act.Account)
		return nil
	}

	// Get message partition key
	partitionKey := s.getPartitionKey(action)
	if partitionKey == "" {
		log.Printf("Invalid partition key for action: %s", action.Act.Name)
		return nil
	}

	if !s.shouldProcessMessage(partitionKey) {
		log.Printf("Skip message with partition key: %s", partitionKey)
		return nil
	}

	// Execute handler
	return handler(action)
}

// getPartitionKey returns partition key based on action type
func (s *Service) getPartitionKey(action hyperion.Action) string {
	switch action.Act.Name {
	case s.eosCfg.Events.EmitPlaced, s.eosCfg.Events.EmitCanceled, s.eosCfg.Events.EmitFilled:
		// Parse action data to get poolID
		var data struct {
			EV struct {
				PoolID string `json:"pool_id"`
			} `json:"ev"`
		}
		if err := json.Unmarshal(action.Act.Data, &data); err != nil {
			log.Printf("Unmarshal action data failed: %v", err)
			return ""
		}
		return fmt.Sprintf("pool-%s", data.EV.PoolID)
	case s.eosCfg.Events.Create:
		// Use fixed partition key for pool creation
		return "pool-creation"
	case s.eosCfg.Events.LogNewAcc, s.eosCfg.Events.DepositLog:
		return fmt.Sprintf("deposit-or-create-account")

	case s.eosCfg.Events.WithdrawLog, s.eosCfg.Events.LogWithdraw:
		return fmt.Sprintf("withdraw")

	default:
		return ""
	}
}

// shouldProcessMessage determines if should process this message
func (s *Service) shouldProcessMessage(partitionKey string) bool {
	// Use consistent hashing to determine whether to process the message
	hasher := fnv.New32a()
	hasher.Write([]byte(partitionKey))
	hash := hasher.Sum32()

	// Get current instance info
	currentInstance, totalInstances := s.getInstanceInfo(context.Background())
	if totalInstances == 0 {
		totalInstances = 1
	}

	// Use hash value to determine which instance should process the message
	targetInstance := int(hash % uint32(totalInstances))
	return targetInstance == currentInstance
}

// initKlineCache initializes the kline cache
func (s *Service) initKlineCache(ctx context.Context) error {
	// Get all trading pairs
	pools, err := s.repo.GetAllPools(ctx)
	if err != nil {
		return fmt.Errorf("get all pools failed: %w", err)
	}

	// Get latest two klines for each trading pair
	for _, pool := range pools {
		klines, err := s.ckhRepo.GetLatestTwoKlines(ctx, pool.PoolID)
		if err != nil {
			log.Printf("get latest two klines failed for pool %d: %v", pool.PoolID, err)
			continue
		}

		// Initialize kline cache for this trading pair
		s.klineCache[pool.PoolID] = make(map[ckhdb.KlineInterval]*ckhdb.Kline)

		// Group kline data by interval
		klineMap := make(map[ckhdb.KlineInterval][]*ckhdb.Kline)
		for _, kline := range klines {
			klineMap[kline.Interval] = append(klineMap[kline.Interval], kline)
		}

		// Process klines for each interval
		for interval, intervalKlines := range klineMap {
			if len(intervalKlines) > 0 {
				// If there are two klines, set the latest kline's open price to previous kline's close price
				if len(intervalKlines) == 2 {
					intervalKlines[0].Open = intervalKlines[1].Close
				}
				s.klineCache[pool.PoolID][interval] = intervalKlines[0] // Only cache the latest kline
			}
		}
	}
	return nil
}
