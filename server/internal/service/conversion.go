package service

import (
	"context"
	"math"

	"fx-tracker/internal/model"
	"fx-tracker/internal/repository"
)

type DCAStatus struct {
	BlendedRate float64 `json:"blended_rate"`
	LiveRate    float64 `json:"live_rate"`
	BeatsAvg    bool    `json:"beats_avg"`
	DeltaPct    float64 `json:"delta_pct"`
	TotalMyr    float64 `json:"total_myr"`
	TotalUsd    float64 `json:"total_usd"`
	HasData     bool    `json:"has_data"`
}

type ConversionService struct {
	repo repository.ConversionRepository
}

func NewConversionService(repo repository.ConversionRepository) *ConversionService {
	return &ConversionService{repo: repo}
}

func (s *ConversionService) Create(ctx context.Context, c *model.Conversion) error {
	return s.repo.Create(ctx, c)
}

func (s *ConversionService) List(ctx context.Context) ([]model.Conversion, error) {
	return s.repo.List(ctx)
}

// Status compares the live rate (USD per MYR) to the blended average. A higher
// live rate means more USD per MYR now than the average cost, i.e. beats_avg.
func (s *ConversionService) Status(ctx context.Context, liveRate float64) (DCAStatus, error) {
	blended, totalMyr, totalUsd, err := s.repo.BlendedRate(ctx)
	if err != nil {
		return DCAStatus{}, err
	}
	if totalMyr == 0 {
		return DCAStatus{LiveRate: liveRate, HasData: false}, nil
	}
	return DCAStatus{
		BlendedRate: blended,
		LiveRate:    liveRate,
		BeatsAvg:    liveRate > blended,
		DeltaPct:    math.Round((liveRate-blended)/blended*100*1e10) / 1e10,
		TotalMyr:    totalMyr,
		TotalUsd:    totalUsd,
		HasData:     true,
	}, nil
}
