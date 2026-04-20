package service

import (
	"context"
	"log/slog"

	"github.com/Timoth26/GopherCart-Bridge/internal/domain"
)

type OrderService struct {
	repo domain.OrderRepository
	log  *slog.Logger
}

func NewOrderService(repo domain.OrderRepository, log *slog.Logger) *OrderService {
	return &OrderService{repo: repo, log: log}
}

func (s *OrderService) GetByID(ctx context.Context, id int64) (*domain.Order, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *OrderService) GetAll(ctx context.Context) ([]domain.Order, error) {
	return s.repo.GetAll(ctx)
}

func (s *OrderService) Create(ctx context.Context, o *domain.Order) error {
	return s.repo.Create(ctx, o)
}

func (s *OrderService) UpdateStatus(ctx context.Context, id int64, status string) error {
	return s.repo.UpdateStatus(ctx, id, status)
}

func (s *OrderService) Delete(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}
