package service

import (
	"context"
	"math"

	"fx-tracker/internal/model"
	"fx-tracker/internal/repository"
)

type DCAStatus struct {
	From        string  `json:"from"`
	To          string  `json:"to"`
	BlendedRate float64 `json:"blended_rate"`
	LiveRate    float64 `json:"live_rate"`
	BeatsAvg    bool    `json:"beats_avg"`
	DeltaPct    float64 `json:"delta_pct"`
	TotalHome   float64 `json:"total_home"`
	TotalTarget float64 `json:"total_target"`
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

func (s *ConversionService) Update(ctx context.Context, c *model.Conversion) error {
	return s.repo.Update(ctx, c)
}

func (s *ConversionService) Delete(ctx context.Context, id uint) error {
	return s.repo.Delete(ctx, id)
}

// Status compares the live rate (target per 1 home) to the blended average
// for a given currency pair. A higher live rate means more target currency
// per unit of home currency now than the average cost, i.e. beats_avg.
func (s *ConversionService) Status(ctx context.Context, from, to string, liveRate float64) (DCAStatus, error) {
	// A non-positive liveRate means we haven't fetched a live FX rate yet
	// (cold start) or the FX source was down at boot — there's nothing to
	// compare the blended average against, so signal "no comparison" rather
	// than a misleading -100% delta.
	if liveRate <= 0 {
		return DCAStatus{From: from, To: to, LiveRate: liveRate, HasData: false}, nil
	}
	blended, totalHome, totalTarget, err := s.repo.BlendedRate(ctx, from, to)
	if err != nil {
		return DCAStatus{}, err
	}
	if totalHome == 0 {
		return DCAStatus{From: from, To: to, LiveRate: liveRate, HasData: false}, nil
	}
	return DCAStatus{
		From:        from,
		To:          to,
		BlendedRate: blended,
		LiveRate:    liveRate,
		BeatsAvg:    liveRate > blended,
		DeltaPct:    math.Round((liveRate-blended)/blended*100*1e10) / 1e10,
		TotalHome:   totalHome,
		TotalTarget: totalTarget,
		HasData:     true,
	}, nil
}
