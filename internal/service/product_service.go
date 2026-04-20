package service

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/Timoth26/GopherCart-Bridge/internal/domain"
)

const productCacheTTL = 5 * time.Minute

type ProductService struct {
	repo  domain.ProductRepository
	cache domain.ProductCache
	log   *slog.Logger
}

func NewProductService(repo domain.ProductRepository, cache domain.ProductCache, log *slog.Logger) *ProductService {
	return &ProductService{repo: repo, cache: cache, log: log}
}

func (s *ProductService) GetByID(ctx context.Context, id int64) (*domain.Product, error) {
	p, err := s.cache.Get(ctx, id)
	if err == nil {
		s.log.DebugContext(ctx, "product cache hit", "id", id)
		return p, nil
	}
	if !errors.Is(err, domain.ErrCacheMiss) {
		s.log.WarnContext(ctx, "product cache get failed", "id", id, "err", err)
	}

	p, err = s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := s.cache.Set(ctx, p, productCacheTTL); err != nil {
		s.log.WarnContext(ctx, "product cache set failed", "id", id, "err", err)
	}

	return p, nil
}

func (s *ProductService) GetAll(ctx context.Context) ([]domain.Product, error) {
	return s.repo.GetAll(ctx)
}

func (s *ProductService) Create(ctx context.Context, p *domain.Product) error {
	if err := s.repo.Create(ctx, p); err != nil {
		return err
	}

	if err := s.cache.Delete(ctx, p.ID); err != nil {
		s.log.WarnContext(ctx, "product cache delete failed after create", "id", p.ID, "err", err)
	}

	return nil
}

func (s *ProductService) Update(ctx context.Context, p *domain.Product) error {
	if err := s.repo.Update(ctx, p); err != nil {
		return err
	}

	if err := s.cache.Delete(ctx, p.ID); err != nil {
		s.log.WarnContext(ctx, "product cache delete failed after update", "id", p.ID, "err", err)
	}

	return nil
}

func (s *ProductService) Delete(ctx context.Context, id int64) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}

	if err := s.cache.Delete(ctx, id); err != nil {
		s.log.WarnContext(ctx, "product cache delete failed after delete", "id", id, "err", err)
	}

	return nil
}
