package service

import (
	"context"

	"fx-tracker/internal/model"
	"fx-tracker/internal/repository"
)

// HoldingService is a thin delegation layer over HoldingRepository, keeping
// handlers off GORM/repository types directly.
type HoldingService struct {
	repo repository.HoldingRepository
}

func NewHoldingService(repo repository.HoldingRepository) *HoldingService {
	return &HoldingService{repo: repo}
}

func (s *HoldingService) Create(ctx context.Context, h *model.Holding) error {
	return s.repo.Create(ctx, h)
}

func (s *HoldingService) List(ctx context.Context) ([]model.Holding, error) {
	return s.repo.List(ctx)
}

func (s *HoldingService) Update(ctx context.Context, h *model.Holding) error {
	return s.repo.Update(ctx, h)
}

func (s *HoldingService) Delete(ctx context.Context, id uint) error {
	return s.repo.Delete(ctx, id)
}
