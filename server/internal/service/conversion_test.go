package service

import (
	"context"
	"errors"
	"testing"

	"fx-tracker/internal/model"
	"fx-tracker/internal/repository"
	"gorm.io/gorm"
)

// fakeRepo implements repository.ConversionRepository in memory.
type fakeRepo struct {
	rate, totalHome, totalTarget float64
	err                          error
	pairs                        []repository.Pair
	updated                      *model.Conversion
	deleted                      uint
}

func (f *fakeRepo) Create(ctx context.Context, c *model.Conversion) error { return f.err }
func (f *fakeRepo) List(ctx context.Context) ([]model.Conversion, error)  { return nil, f.err }
func (f *fakeRepo) BlendedRate(ctx context.Context, from, to string) (float64, float64, float64, error) {
	return f.rate, f.totalHome, f.totalTarget, f.err
}
func (f *fakeRepo) Pairs(ctx context.Context) ([]repository.Pair, error) {
	return f.pairs, f.err
}
func (f *fakeRepo) Update(ctx context.Context, c *model.Conversion) error {
	f.updated = c
	return f.err
}
func (f *fakeRepo) Delete(ctx context.Context, id uint) error { f.deleted = id; return f.err }

func TestStatus(t *testing.T) {
	tests := []struct {
		name         string
		repo         *fakeRepo
		from, to     string
		live         float64
		wantHasData  bool
		wantBeats    bool
		wantDeltaPct float64
	}{
		{"no conversions", &fakeRepo{}, "MYR", "USD", 0.24, false, false, 0},
		{"live beats average", &fakeRepo{rate: 0.20, totalHome: 1000, totalTarget: 200}, "MYR", "USD", 0.205, true, true, 2.5},
		{"live below average", &fakeRepo{rate: 0.20, totalHome: 1000, totalTarget: 200}, "MYR", "USD", 0.195, true, false, -2.5},
		{"live rate unavailable", &fakeRepo{rate: 0.20, totalHome: 1000, totalTarget: 200}, "MYR", "USD", 0, false, false, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewConversionService(tt.repo)
			got, err := s.Status(context.Background(), tt.from, tt.to, tt.live)
			if err != nil {
				t.Fatal(err)
			}
			if got.From != tt.from || got.To != tt.to {
				t.Fatalf("pair = %s/%s, want %s/%s", got.From, got.To, tt.from, tt.to)
			}
			if got.HasData != tt.wantHasData || got.BeatsAvg != tt.wantBeats {
				t.Fatalf("status = %+v", got)
			}
			if got.HasData && got.DeltaPct != tt.wantDeltaPct {
				t.Fatalf("delta = %v, want %v", got.DeltaPct, tt.wantDeltaPct)
			}
		})
	}
}

func TestUpdateDeleteDelegate(t *testing.T) {
	repo := &fakeRepo{}
	s := NewConversionService(repo)
	ctx := context.Background()

	c := &model.Conversion{ID: 7, FromCurrency: "MYR", ToCurrency: "USD", FromAmount: 100, Rate: 0.2}
	if err := s.Update(ctx, c); err != nil {
		t.Fatal(err)
	}
	if repo.updated == nil || repo.updated.ID != 7 {
		t.Fatalf("update not delegated: %+v", repo.updated)
	}
	if err := s.Delete(ctx, 7); err != nil {
		t.Fatal(err)
	}
	if repo.deleted != 7 {
		t.Fatalf("delete not delegated: %d", repo.deleted)
	}

	// error passthrough (e.g. not found)
	errRepo := &fakeRepo{err: gorm.ErrRecordNotFound}
	s2 := NewConversionService(errRepo)
	if err := s2.Delete(ctx, 1); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("passthrough: got %v", err)
	}
}
