package repository

import (
	"context"

	"gorm.io/gorm"

	"fx-tracker/internal/model"
)

type ConversionRepository interface {
	Create(ctx context.Context, c *model.Conversion) error
	List(ctx context.Context) ([]model.Conversion, error)
	// BlendedRate returns total USD acquired / total MYR spent (USD per MYR),
	// plus the totals. Zeros (no error) when there are no conversions.
	BlendedRate(ctx context.Context) (rate, totalMyr, totalUsd float64, err error)
}

type conversionRepo struct {
	db *gorm.DB
}

func NewConversionRepository(db *gorm.DB) ConversionRepository {
	return &conversionRepo{db: db}
}

func (r *conversionRepo) Create(ctx context.Context, c *model.Conversion) error {
	return r.db.WithContext(ctx).Create(c).Error
}

func (r *conversionRepo) List(ctx context.Context) ([]model.Conversion, error) {
	var out []model.Conversion
	err := r.db.WithContext(ctx).Order("date DESC, id DESC").Find(&out).Error
	return out, err
}

func (r *conversionRepo) BlendedRate(ctx context.Context) (rate, totalMyr, totalUsd float64, err error) {
	var res struct {
		TotalUsd float64
		TotalMyr float64
	}
	err = r.db.WithContext(ctx).Model(&model.Conversion{}).
		Select("COALESCE(SUM(myr_amount * rate_myr_usd), 0) AS total_usd, COALESCE(SUM(myr_amount), 0) AS total_myr").
		Scan(&res).Error
	if err != nil {
		return 0, 0, 0, err
	}
	if res.TotalMyr > 0 {
		rate = res.TotalUsd / res.TotalMyr
	}
	return rate, res.TotalMyr, res.TotalUsd, nil
}
