package worker

import (
	"context"
	"errors"
	"log/slog"
	"sync"

	"github.com/Timoth26/GopherCart-Bridge/internal/domain"
)

var ErrPoolStopped = errors.New("worker pool stopped")

type productService interface {
	Upsert(ctx context.Context, p *domain.Product) error
}

type orderService interface {
	GetPending(ctx context.Context) ([]domain.Order, error)
	UpdateStatus(ctx context.Context, id int64, status string) error
}

type JobKind string

const (
	SyncJob  JobKind = "sync_job"
	OrderJob JobKind = "order_job"
)

type Job struct {
	Kind JobKind
}

type Pool struct {
	jobs     chan Job
	done     chan struct{}
	supplier domain.SupplierClient
	products productService
	orders   orderService
	workers  int
	log      *slog.Logger
	wg       sync.WaitGroup
}

func NewPool(
	workers int,
	bufSize int,
	supplier domain.SupplierClient,
	products productService,
	orders orderService,
	log *slog.Logger,
) *Pool {
	return &Pool{
		jobs:     make(chan Job, bufSize),
		done:     make(chan struct{}),
		supplier: supplier,
		products: products,
		orders:   orders,
		workers:  workers,
		log:      log,
	}
}

func (p *Pool) Start(ctx context.Context) {
	for i := 0; i < p.workers; i++ {
		p.wg.Add(1)
		go p.work(ctx, i)
	}
}

func (p *Pool) Submit(ctx context.Context, job Job) error {
	select {
	case p.jobs <- job:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	case <-p.done:
		return ErrPoolStopped
	}
}

func (p *Pool) Stop() {
	close(p.done)
	p.wg.Wait()
}

func (p *Pool) work(ctx context.Context, id int) {
	defer p.wg.Done()
	for {
		select {
		case job := <-p.jobs:
			p.handle(ctx, id, job)
		case <-ctx.Done():
			return
		case <-p.done:
			return
		}
	}
}

func (p *Pool) handle(ctx context.Context, workerID int, job Job) {
	p.log.InfoContext(ctx, "processing job", "worker", workerID, "kind", job.Kind)

	var err error
	switch job.Kind {
	case SyncJob:
		err = newSyncJob(p.supplier, p.products, p.log).run(ctx)
	case OrderJob:
		err = newOrderJob(p.supplier, p.orders, p.log).run(ctx)
	default:
		p.log.WarnContext(ctx, "unknown job kind", "worker", workerID, "kind", job.Kind)
		return
	}

	if err != nil {
		p.log.ErrorContext(ctx, "job failed", "worker", workerID, "kind", job.Kind, "err", err)
		return
	}

	p.log.InfoContext(ctx, "job done", "worker", workerID, "kind", job.Kind)
}
