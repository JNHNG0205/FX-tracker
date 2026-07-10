package service

import (
	"context"
	"testing"

	"fx-tracker/internal/model"
)

// fakeRepo implements repository.ConversionRepository in memory.
type fakeRepo struct {
	rate, totalMyr, totalUsd float64
	err                      error
}

func (f *fakeRepo) Create(ctx context.Context, c *model.Conversion) error { return f.err }
func (f *fakeRepo) List(ctx context.Context) ([]model.Conversion, error)  { return nil, f.err }
func (f *fakeRepo) BlendedRate(ctx context.Context) (float64, float64, float64, error) {
	return f.rate, f.totalMyr, f.totalUsd, f.err
}

func TestStatus(t *testing.T) {
	tests := []struct {
		name         string
		repo         *fakeRepo
		live         float64
		wantHasData  bool
		wantBeats    bool
		wantDeltaPct float64
	}{
		{"no conversions", &fakeRepo{}, 0.24, false, false, 0},
		{"live beats average", &fakeRepo{rate: 0.20, totalMyr: 1000, totalUsd: 200}, 0.205, true, true, 2.5},
		{"live below average", &fakeRepo{rate: 0.20, totalMyr: 1000, totalUsd: 200}, 0.195, true, false, -2.5},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewConversionService(tt.repo)
			got, err := s.Status(context.Background(), tt.live)
			if err != nil {
				t.Fatal(err)
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
