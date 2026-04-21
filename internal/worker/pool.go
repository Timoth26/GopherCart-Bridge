package worker

import (
	"context"
	"log/slog"
	"sync"

	"github.com/Timoth26/GopherCart-Bridge/internal/domain"
)

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
	supplier domain.SupplierClient
	products productService
	cache    domain.ProductCache
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
	cache domain.ProductCache,
	orders orderService,
	log *slog.Logger,
) *Pool {
	return &Pool{
		jobs:     make(chan Job, bufSize),
		supplier: supplier,
		products: products,
		cache:    cache,
		orders:   orders,
		workers:  workers,
		log:      log,
	}
}

func (p *Pool) Start(ctx context.Context) {
	for i := range p.workers {
		p.wg.Add(1)
		go p.work(ctx, i)
	}
}

func (p *Pool) Submit(job Job) {
	p.jobs <- job
}

func (p *Pool) Stop() {
	close(p.jobs)
	p.wg.Wait()
}

func (p *Pool) work(ctx context.Context, id int) {
	defer p.wg.Done()
	for {
		select {
		case job, ok := <-p.jobs:
			if !ok {
				return
			}
			p.handle(ctx, id, job)
		case <-ctx.Done():
			return
		}
	}
}

func (p *Pool) handle(ctx context.Context, workerID int, job Job) {
	p.log.InfoContext(ctx, "processing job", "worker", workerID, "kind", job.Kind)

	var err error
	switch job.Kind {
	case SyncJob:
		err = newSyncJob(p.supplier, p.products, p.cache, p.log).run(ctx)
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
