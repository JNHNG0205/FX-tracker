package repository

import (
	"context"

	"gorm.io/gorm"

	"fx-tracker/internal/model"
)

type HoldingRepository interface {
	Create(ctx context.Context, h *model.Holding) error
	List(ctx context.Context) ([]model.Holding, error)
	Update(ctx context.Context, h *model.Holding) error
	Delete(ctx context.Context, id uint) error
}

type holdingRepo struct {
	db *gorm.DB
}

func NewHoldingRepository(db *gorm.DB) HoldingRepository {
	return &holdingRepo{db: db}
}

func (r *holdingRepo) Create(ctx context.Context, h *model.Holding) error {
	return r.db.WithContext(ctx).Create(h).Error
}

func (r *holdingRepo) List(ctx context.Context) ([]model.Holding, error) {
	var out []model.Holding
	err := r.db.WithContext(ctx).Order("id DESC").Find(&out).Error
	return out, err
}

// Update explicitly Selects the updatable columns, including manual_price,
// so struct-based Updates still writes it when it's nil/zero — GORM's
// default Updates skips zero-valued fields, which would otherwise make a
// cleared manual_price (nil pointer) silently fail to persist.
func (r *holdingRepo) Update(ctx context.Context, h *model.Holding) error {
	res := r.db.WithContext(ctx).Model(&model.Holding{}).
		Where("id = ?", h.ID).
		Select("ticker", "shares", "avg_cost", "currency", "manual_price").
		Updates(h)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *holdingRepo) Delete(ctx context.Context, id uint) error {
	res := r.db.WithContext(ctx).Delete(&model.Holding{}, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
