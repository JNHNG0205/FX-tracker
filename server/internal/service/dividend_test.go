package service

import (
	"context"
	"testing"
	"time"

	"github.com/shopspring/decimal"

	"fx-tracker/internal/fx"
	"fx-tracker/internal/model"
)

type fakeDividendRepo struct {
	dividends []model.Dividend
}

func (f *fakeDividendRepo) Create(ctx context.Context, d *model.Dividend) error {
	f.dividends = append(f.dividends, *d)
	return nil
}

func (f *fakeDividendRepo) List(ctx context.Context) ([]model.Dividend, error) {
	return f.dividends, nil
}

func (f *fakeDividendRepo) Update(ctx context.Context, d *model.Dividend) error {
	return nil
}

func (f *fakeDividendRepo) Delete(ctx context.Context, id uint) error {
	return nil
}

type fakeDividendSpotSource struct {
	rates map[string]float64
	errs  map[string]bool
}

func (f *fakeDividendSpotSource) Rate(ctx context.Context, from, to string) (fx.Rate, error) {
	key := from + ">" + to
	if f.errs[key] {
		return fx.Rate{}, context.DeadlineExceeded
	}
	return fx.Rate{From: from, To: to, Rate: f.rates[key]}, nil
}

func mustDividendDate(s string) time.Time {
	t, _ := time.Parse("2006-01-02", s)
	return t
}

func TestDividendServiceSummary(t *testing.T) {
	repo := &fakeDividendRepo{dividends: []model.Dividend{
		{ID: 1, Ticker: "VOO", Currency: "USD", Amount: decimal.RequireFromString("100.00"), Date: mustDividendDate("2026-01-15"), Note: "q1"},
		{ID: 2, Ticker: "EUFUND", Currency: "EUR", Amount: decimal.RequireFromString("50.00"), Date: mustDividendDate("2026-02-15"), Note: "q2"},
		{ID: 3, Ticker: "MYRFUND", Currency: "MYR", Amount: decimal.RequireFromString("20.00"), Date: mustDividendDate("2026-03-15"), Note: "q3"},
	}}
	spot := &fakeDividendSpotSource{
		rates: map[string]float64{"MYR>USD": 0.22},
		errs:  map[string]bool{"MYR>EUR": true},
	}

	svc := NewDividendService(repo, spot)
	summary, err := svc.Summary(context.Background(), "MYR")
	if err != nil {
		t.Fatalf("Summary() error = %v", err)
	}
	if summary.HomeCurrency != "MYR" {
		t.Fatalf("HomeCurrency = %s, want MYR", summary.HomeCurrency)
	}
	if len(summary.Dividends) != 3 {
		t.Fatalf("expected 3 dividends, got %d", len(summary.Dividends))
	}

	usd := summary.Dividends[0]
	if usd.Ticker != "VOO" {
		t.Fatalf("expected first dividend VOO, got %s", usd.Ticker)
	}
	if usd.Withholding != "30" {
		t.Fatalf("Withholding = %s, want 30", usd.Withholding)
	}
	if usd.Net != "70" {
		t.Fatalf("Net = %s, want 70", usd.Net)
	}
	if !usd.HomeAvailable {
		t.Fatalf("expected VOO HomeAvailable true")
	}
	wantHomeNet := 70.0 / 0.22
	if diff := usd.HomeNet - wantHomeNet; diff > 1e-9 || diff < -1e-9 {
		t.Fatalf("HomeNet = %v, want %v", usd.HomeNet, wantHomeNet)
	}
	if usd.Date != "2026-01-15" {
		t.Fatalf("Date = %s, want 2026-01-15", usd.Date)
	}

	eur := summary.Dividends[1]
	if eur.Withholding != "0" {
		t.Fatalf("Withholding = %s, want 0", eur.Withholding)
	}
	if eur.HomeAvailable {
		t.Fatalf("expected EUR dividend HomeAvailable false (spot error)")
	}
	if eur.HomeNet != 0 {
		t.Fatalf("expected HomeNet = 0 for unavailable entry, got %v", eur.HomeNet)
	}

	myr := summary.Dividends[2]
	if !myr.HomeAvailable {
		t.Fatalf("expected home==currency dividend HomeAvailable true")
	}
	if myr.HomeNet != 20.0 {
		t.Fatalf("HomeNet = %v, want 20.0 for home==currency", myr.HomeNet)
	}

	if len(summary.Totals.ByCurrency) != 3 {
		t.Fatalf("expected 3 currency totals, got %d", len(summary.Totals.ByCurrency))
	}
	byCurrency := make(map[string]CurrencyTotal, len(summary.Totals.ByCurrency))
	for _, ct := range summary.Totals.ByCurrency {
		byCurrency[ct.Currency] = ct
	}
	if byCurrency["USD"].Gross != "100" || byCurrency["USD"].Withholding != "30" || byCurrency["USD"].Net != "70" {
		t.Fatalf("USD total = %+v", byCurrency["USD"])
	}
	if byCurrency["EUR"].Gross != "50" || byCurrency["EUR"].Withholding != "0" || byCurrency["EUR"].Net != "50" {
		t.Fatalf("EUR total = %+v", byCurrency["EUR"])
	}
	if byCurrency["MYR"].Gross != "20" {
		t.Fatalf("MYR total = %+v", byCurrency["MYR"])
	}

	// EUR home_net excluded because spot lookup errored.
	wantTotalHomeNet := wantHomeNet + 20.0
	if diff := summary.Totals.HomeNet - wantTotalHomeNet; diff > 1e-9 || diff < -1e-9 {
		t.Fatalf("Totals.HomeNet = %v, want %v", summary.Totals.HomeNet, wantTotalHomeNet)
	}
}
