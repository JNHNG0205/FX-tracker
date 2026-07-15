package repository

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"fx-tracker/internal/model"
)

// SettingsRepository manages the single-row app settings table (fixed ID 1).
type SettingsRepository interface {
	Get(ctx context.Context) (model.Setting, error)
	SetHomeCurrency(ctx context.Context, code string) error
}

type settingsRepo struct {
	db *gorm.DB
}

func NewSettingsRepository(db *gorm.DB) SettingsRepository {
	return &settingsRepo{db: db}
}

// Get returns the single settings row, creating it with the default
// HomeCurrency "MYR" if it doesn't exist yet — callers never see an error
// just because settings haven't been configured.
func (r *settingsRepo) Get(ctx context.Context) (model.Setting, error) {
	var s model.Setting
	err := r.db.WithContext(ctx).
		Attrs(model.Setting{HomeCurrency: "MYR"}).
		FirstOrCreate(&s, model.Setting{ID: 1}).Error
	return s, err
}

// SetHomeCurrency upserts the single settings row (ID 1) with the given
// currency code.
func (r *settingsRepo) SetHomeCurrency(ctx context.Context, code string) error {
	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "id"}},
			DoUpdates: clause.AssignmentColumns([]string{"home_currency"}),
		}).
		Create(&model.Setting{ID: 1, HomeCurrency: code}).Error
}
