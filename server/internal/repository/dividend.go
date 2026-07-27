package repository

import (
	"context"

	"gorm.io/gorm"

	"fx-tracker/internal/model"
)

type DividendRepository interface {
	Create(ctx context.Context, d *model.Dividend) error
	List(ctx context.Context) ([]model.Dividend, error)
	Update(ctx context.Context, d *model.Dividend) error
	Delete(ctx context.Context, id uint) error
}

type dividendRepo struct {
	db *gorm.DB
}

func NewDividendRepository(db *gorm.DB) DividendRepository {
	return &dividendRepo{db: db}
}

func (r *dividendRepo) Create(ctx context.Context, d *model.Dividend) error {
	return r.db.WithContext(ctx).Create(d).Error
}

func (r *dividendRepo) List(ctx context.Context) ([]model.Dividend, error) {
	var out []model.Dividend
	err := r.db.WithContext(ctx).Order("date DESC, id DESC").Find(&out).Error
	return out, err
}

func (r *dividendRepo) Update(ctx context.Context, d *model.Dividend) error {
	res := r.db.WithContext(ctx).Model(&model.Dividend{}).
		Where("id = ?", d.ID).
		Select("ticker", "currency", "amount", "date", "note").
		Updates(d)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *dividendRepo) Delete(ctx context.Context, id uint) error {
	res := r.db.WithContext(ctx).Delete(&model.Dividend{}, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
