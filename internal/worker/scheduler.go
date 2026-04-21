package worker

import (
	"context"
	"log/slog"
	"time"
)

type Scheduler struct {
	pool          *Pool
	syncInterval  time.Duration
	orderInterval time.Duration
	log           *slog.Logger
}

func NewScheduler(pool *Pool, syncInterval, orderInterval time.Duration, log *slog.Logger) *Scheduler {
	return &Scheduler{
		pool:          pool,
		syncInterval:  syncInterval,
		orderInterval: orderInterval,
		log:           log,
	}
}

func (s *Scheduler) Run(ctx context.Context) {
	syncTicker := time.NewTicker(s.syncInterval)
	orderTicker := time.NewTicker(s.orderInterval)
	defer syncTicker.Stop()
	defer orderTicker.Stop()

	s.log.InfoContext(ctx, "scheduler started",
		"sync_interval", s.syncInterval,
		"order_interval", s.orderInterval,
	)

	for {
		select {
		case <-syncTicker.C:
			s.log.InfoContext(ctx, "submitting sync job")
			s.pool.Submit(Job{Kind: SyncJob})

		case <-orderTicker.C:
			s.log.InfoContext(ctx, "submitting order job")
			s.pool.Submit(Job{Kind: OrderJob})

		case <-ctx.Done():
			s.log.InfoContext(ctx, "scheduler stopped")
			return
		}
	}
}
