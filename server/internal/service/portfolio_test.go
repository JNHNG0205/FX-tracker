package service

import (
	"context"
	"testing"

	"fx-tracker/internal/fx"
	"fx-tracker/internal/model"
	"fx-tracker/internal/price"
)

type fakeHoldingLister struct {
	holdings []model.Holding
}

func (f *fakeHoldingLister) List(ctx context.Context) ([]model.Holding, error) {
	return f.holdings, nil
}

type fakePriceSource struct {
	quotes map[string]price.Quote
}

func (f *fakePriceSource) Quotes(ctx context.Context, reqs []price.Req) map[string]price.Quote {
	out := make(map[string]price.Quote, len(reqs))
	for _, r := range reqs {
		if r.Manual != nil {
			out[r.Ticker] = price.Quote{Ticker: r.Ticker, Price: *r.Manual, Found: true}
			continue
		}
		if q, ok := f.quotes[r.Ticker]; ok {
			out[r.Ticker] = q
		}
	}
	return out
}

type fakeBlendedSource struct {
	rates map[string]float64 // key: from+">"+to
}

func (f *fakeBlendedSource) BlendedRate(ctx context.Context, from, to string) (float64, float64, float64, error) {
	rate := f.rates[from+">"+to]
	if rate == 0 {
		return 0, 0, 0, nil
	}
	return rate, 100, 100 * rate, nil
}

type fakeSpotSource struct {
	rates map[string]float64
}

func (f *fakeSpotSource) Rate(ctx context.Context, from, to string) (fx.Rate, error) {
	return fx.Rate{From: from, To: to, Rate: f.rates[from+">"+to]}, nil
}

func TestPortfolioServiceCompute(t *testing.T) {
	holdings := []model.Holding{
		{ID: 1, Ticker: "VOO", Shares: 10, AvgCost: 400, Currency: "USD"},
		{ID: 2, Ticker: "EUFUND", Shares: 5, AvgCost: 100, Currency: "EUR"},
	}

	holdingLister := &fakeHoldingLister{holdings: holdings}
	prices := &fakePriceSource{quotes: map[string]price.Quote{
		"VOO":    {Ticker: "VOO", Price: 500, Found: true},
		"EUFUND": {Ticker: "EUFUND", Found: false},
	}}
	blended := &fakeBlendedSource{rates: map[string]float64{
		"MYR>USD": 0.22,
	}}
	spot := &fakeSpotSource{rates: map[string]float64{
		"MYR>USD": 0.24,
		"MYR>EUR": 0.20,
	}}

	svc := NewPortfolioService(holdingLister, prices, blended, spot)

	resp, err := svc.Compute(context.Background(), "MYR")
	if err != nil {
		t.Fatalf("Compute() error = %v", err)
	}
	if len(resp.Holdings) != 2 {
		t.Fatalf("expected 2 holdings, got %d", len(resp.Holdings))
	}

	usd := resp.Holdings[0]
	if usd.Ticker != "VOO" {
		t.Fatalf("expected first holding VOO, got %s", usd.Ticker)
	}
	if !usd.HomeAvailable {
		t.Fatalf("expected VOO holding to be HomeAvailable")
	}
	if usd.PriceSource != "finnhub" {
		t.Fatalf("expected price source finnhub, got %s", usd.PriceSource)
	}
	wantCost := 10 * 400.0
	wantValue := 10 * 500.0
	wantHomeCost := wantCost / 0.22
	wantHomeValue := wantValue / 0.24
	wantTotalPct := (wantHomeValue - wantHomeCost) / wantHomeCost * 100
	if diff := usd.TotalReturnPct - wantTotalPct; diff > 1e-9 || diff < -1e-9 {
		t.Fatalf("TotalReturnPct = %v, want %v", usd.TotalReturnPct, wantTotalPct)
	}

	eur := resp.Holdings[1]
	if eur.Ticker != "EUFUND" {
		t.Fatalf("expected second holding EUFUND, got %s", eur.Ticker)
	}
	if eur.HomeAvailable {
		t.Fatalf("expected EUFUND holding to be degraded (not HomeAvailable)")
	}
	if eur.PriceSource != "unavailable" {
		t.Fatalf("expected price source unavailable, got %s", eur.PriceSource)
	}

	if resp.Totals.Counted != 1 {
		t.Fatalf("expected Totals.Counted = 1, got %d", resp.Totals.Counted)
	}
	if diff := resp.Totals.HomeCost - wantHomeCost; diff > 1e-9 || diff < -1e-9 {
		t.Fatalf("Totals.HomeCost = %v, want %v", resp.Totals.HomeCost, wantHomeCost)
	}
}

func TestPortfolioServiceComputeSameCurrencyAsHome(t *testing.T) {
	holdings := []model.Holding{
		{ID: 1, Ticker: "VOO", Shares: 10, AvgCost: 400, Currency: "USD"},
	}
	holdingLister := &fakeHoldingLister{holdings: holdings}
	prices := &fakePriceSource{quotes: map[string]price.Quote{
		"VOO": {Ticker: "VOO", Price: 500, Found: true},
	}}
	// Deliberately empty: BlendedRate/Rate for USD>USD must never be called
	// (and would return zero/error if it were), proving the same-currency
	// shortcut bypasses the lookups entirely.
	blended := &fakeBlendedSource{rates: map[string]float64{}}
	spot := &fakeSpotSource{rates: map[string]float64{}}

	svc := NewPortfolioService(holdingLister, prices, blended, spot)
	resp, err := svc.Compute(context.Background(), "USD")
	if err != nil {
		t.Fatalf("Compute() error = %v", err)
	}
	if len(resp.Holdings) != 1 {
		t.Fatalf("expected 1 holding, got %d", len(resp.Holdings))
	}

	h := resp.Holdings[0]
	if !h.HomeAvailable {
		t.Fatalf("expected HomeAvailable true for home==currency holding")
	}
	if h.FxPct != 0 {
		t.Fatalf("expected FxPct = 0, got %v", h.FxPct)
	}
	wantPct := (5000.0 - 4000.0) / 4000.0 * 100
	if diff := h.AssetPnlPct - wantPct; diff > 1e-9 || diff < -1e-9 {
		t.Fatalf("AssetPnlPct = %v, want %v", h.AssetPnlPct, wantPct)
	}
	if diff := h.TotalReturnPct - h.AssetPnlPct; diff > 1e-9 || diff < -1e-9 {
		t.Fatalf("TotalReturnPct = %v, want equal to AssetPnlPct %v", h.TotalReturnPct, h.AssetPnlPct)
	}
}

func TestPortfolioServiceComputeManualPrice(t *testing.T) {
	manual := 42.0
	holdings := []model.Holding{
		{ID: 1, Ticker: "PRIVATE", Shares: 2, AvgCost: 10, Currency: "USD", ManualPrice: &manual},
	}
	holdingLister := &fakeHoldingLister{holdings: holdings}
	prices := &fakePriceSource{quotes: map[string]price.Quote{}}
	blended := &fakeBlendedSource{rates: map[string]float64{"MYR>USD": 0.22}}
	spot := &fakeSpotSource{rates: map[string]float64{"MYR>USD": 0.24}}

	svc := NewPortfolioService(holdingLister, prices, blended, spot)
	resp, err := svc.Compute(context.Background(), "MYR")
	if err != nil {
		t.Fatalf("Compute() error = %v", err)
	}
	if resp.Holdings[0].PriceSource != "manual" {
		t.Fatalf("expected price source manual, got %s", resp.Holdings[0].PriceSource)
	}
	if resp.Holdings[0].Price != manual {
		t.Fatalf("expected price %v, got %v", manual, resp.Holdings[0].Price)
	}
}
