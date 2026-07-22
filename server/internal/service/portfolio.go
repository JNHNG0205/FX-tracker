package service

import (
	"context"

	"fx-tracker/internal/fx"
	"fx-tracker/internal/model"
	"fx-tracker/internal/portfolio"
	"fx-tracker/internal/price"
)

// HoldingLister is the read side of HoldingRepository that PortfolioService
// depends on; the real HoldingRepository satisfies it structurally.
type HoldingLister interface {
	List(ctx context.Context) ([]model.Holding, error)
}

// PriceSource resolves a batch of ticker lookups; price.Cache satisfies it
// structurally.
type PriceSource interface {
	Quotes(ctx context.Context, reqs []price.Req) map[string]price.Quote
}

// BlendedSource returns the cost-weighted average rate paid for a currency
// pair; ConversionRepository satisfies it structurally.
type BlendedSource interface {
	BlendedRate(ctx context.Context, from, to string) (rate, totalHome, totalTarget float64, err error)
}

// SpotSource returns the current live rate for a currency pair; fx.Cache
// satisfies it structurally.
type SpotSource interface {
	Rate(ctx context.Context, from, to string) (fx.Rate, error)
}

// PortfolioService orchestrates holdings, live/manual prices, and FX
// (blended + spot) into the pure portfolio math.
type PortfolioService struct {
	holdings HoldingLister
	prices   PriceSource
	blended  BlendedSource
	spot     SpotSource
}

func NewPortfolioService(holdings HoldingLister, prices PriceSource, blended BlendedSource, spot SpotSource) *PortfolioService {
	return &PortfolioService{holdings: holdings, prices: prices, blended: blended, spot: spot}
}

// Compute builds the full portfolio return breakdown against home currency.
// Only a holdings-list error is fatal; a per-holding blended/spot lookup
// failure degrades that holding (HomeAvailable false) rather than failing
// the whole response.
func (s *PortfolioService) Compute(ctx context.Context, home string) (portfolio.Response, error) {
	holdings, err := s.holdings.List(ctx)
	if err != nil {
		return portfolio.Response{}, err
	}

	reqs := make([]price.Req, len(holdings))
	for i, h := range holdings {
		reqs[i] = price.Req{Ticker: h.Ticker, Currency: h.Currency, Manual: h.ManualPrice}
	}
	quotes := s.prices.Quotes(ctx, reqs)

	results := make([]portfolio.HoldingResult, len(holdings))
	for i, h := range holdings {
		var p float64
		var source string
		if h.ManualPrice != nil {
			p = *h.ManualPrice
			source = "manual"
		} else if q, ok := quotes[h.Ticker]; ok && q.Found {
			p = q.Price
			source = "stooq"
		} else {
			p = 0
			source = "unavailable"
		}

		var blended, spot float64
		if home == h.Currency {
			// Same currency: no conversion happened, so the rate is
			// trivially 1 and the holding's return is pure asset return
			// with zero FX component. Skip the blended/spot lookups
			// entirely — frankfurter.app errors on ?from=X&to=X.
			blended, spot = 1, 1
		} else {
			if rate, _, _, err := s.blended.BlendedRate(ctx, home, h.Currency); err == nil {
				blended = rate
			}
			if r, err := s.spot.Rate(ctx, home, h.Currency); err == nil {
				spot = r.Rate
			}
		}

		results[i] = portfolio.ComputeHolding(h.ID, h.Ticker, h.Currency, h.Shares, h.AvgCost, p, source, blended, spot)
	}

	totals := portfolio.Aggregate(results)
	return portfolio.Response{HomeCurrency: home, Holdings: results, Totals: totals}, nil
}
