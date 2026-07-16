package repository

import (
	"context"

	"gorm.io/gorm"

	"fx-tracker/internal/model"
)

// Pair identifies a distinct home→target currency pair that has at least
// one conversion recorded.
type Pair struct {
	From string
	To   string
}

type ConversionRepository interface {
	Create(ctx context.Context, c *model.Conversion) error
	List(ctx context.Context) ([]model.Conversion, error)
	// BlendedRate returns total target acquired / total home spent for a
	// given currency pair (target per 1 home), plus the totals. Zeros (no
	// error) when there are no conversions for that pair.
	BlendedRate(ctx context.Context, from, to string) (rate, totalHome, totalTarget float64, err error)
	// Pairs returns the distinct (from, to) currency pairs present across
	// all conversions, ordered by from then to.
	Pairs(ctx context.Context) ([]Pair, error)
	Update(ctx context.Context, c *model.Conversion) error
	Delete(ctx context.Context, id uint) error
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

func (r *conversionRepo) BlendedRate(ctx context.Context, from, to string) (rate, totalHome, totalTarget float64, err error) {
	var res struct {
		TotalTarget float64
		TotalHome   float64
	}
	err = r.db.WithContext(ctx).Model(&model.Conversion{}).
		Where("from_currency = ? AND to_currency = ?", from, to).
		Select("COALESCE(SUM(from_amount * rate), 0) AS total_target, COALESCE(SUM(from_amount), 0) AS total_home").
		Scan(&res).Error
	if err != nil {
		return 0, 0, 0, err
	}
	if res.TotalHome > 0 {
		rate = res.TotalTarget / res.TotalHome
	}
	return rate, res.TotalHome, res.TotalTarget, nil
}

func (r *conversionRepo) Pairs(ctx context.Context) ([]Pair, error) {
	var rows []struct {
		FromCurrency string
		ToCurrency   string
	}
	err := r.db.WithContext(ctx).Model(&model.Conversion{}).
		Distinct("from_currency", "to_currency").
		Order("from_currency, to_currency").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]Pair, len(rows))
	for i, row := range rows {
		out[i] = Pair{From: row.FromCurrency, To: row.ToCurrency}
	}
	return out, nil
}

// Update uses a map (not a struct) so zero-valued fields — e.g. an emptied
// note — still persist; struct-based Updates skips zero values.
func (r *conversionRepo) Update(ctx context.Context, c *model.Conversion) error {
	res := r.db.WithContext(ctx).Model(&model.Conversion{}).
		Where("id = ?", c.ID).
		Updates(map[string]any{
			"date":          c.Date,
			"from_currency": c.FromCurrency,
			"to_currency":   c.ToCurrency,
			"from_amount":   c.FromAmount,
			"rate":          c.Rate,
			"note":          c.Note,
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *conversionRepo) Delete(ctx context.Context, id uint) error {
	res := r.db.WithContext(ctx).Delete(&model.Conversion{}, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
