package cron

import (
	"context"
	"exapp-go/internal/entity"
	"log"
)

func (s *Service) SyncPoolStats() {
	log.Println("begin sync pool stats...")
	ctx := context.Background()
	err := s.ckhdb.UpdatePoolStats(ctx)
	if err != nil {
		log.Println("failed to update pool stats", err)
	}
	err = s.ckhdb.OptimizePoolStats(ctx)
	if err != nil {
		log.Println("failed to optimize pool stats", err)
	}
	log.Println("sync pool stats done")
}

func (s *Service) SyncAndBroadcastPoolStats() {
	log.Println("begin sync and broadcast pool stats...")
	ctx := context.Background()

	stats, err := s.ckhdb.ListPoolStats(ctx)
	if err != nil {
		log.Printf("failed to list pool stats: %v\n", err)
		return
	}
	poolStatsMap := make(map[uint64]entity.PoolStats)
	for _, stat := range stats {

		poolStats := entity.PoolStatusFromDB(stat)
		if oldPoolStats, ok := poolStatsMap[poolStats.PoolID]; ok {
			if oldPoolStats.Turnover == poolStats.Turnover {
				continue
			} else {
				poolStatsMap[poolStats.PoolID] = *poolStats
			}
		} else {
			poolStatsMap[poolStats.PoolID] = *poolStats
		}
		msg := struct {
			Type string      `json:"type"`
			Data interface{} `json:"data"`
		}{
			Type: "pool_stats_update",
			Data: poolStats,
		}

		err = s.nsqPublisher.Publish("cdex_updates", msg)
		if err != nil {
			log.Printf("failed to publish pool stats message: %v\n", err)
			continue
		}
	}

	log.Println("sync and broadcast pool stats done")
}
