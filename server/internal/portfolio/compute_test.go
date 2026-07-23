package portfolio

import "testing"

func approx(a, b float64) bool {
	d := a - b
	if d < 0 {
		d = -d
	}
	return d < 1e-6
}

func TestComputeHolding(t *testing.T) {
	// 10 sh @500 USD; price 560; blended 0.21 (USD/MYR cost), spot 0.204 (now).
	r := ComputeHolding(1, "VOO", "USD", 10, 500, 560, "finnhub", 0.21, 0.204)
	if r.CostC != 5000 || r.ValueC != 5600 {
		t.Fatalf("C side: %+v", r)
	}
	if !approx(r.AssetPnlPct, 12.0) {
		t.Fatalf("asset%%=%v", r.AssetPnlPct) // (5600-5000)/5000
	}
	if !approx(r.HomeCost, 5000/0.21) || !approx(r.HomeValue, 5600/0.204) {
		t.Fatalf("home: %+v", r)
	}
	if !approx(r.FxPct, (0.21/0.204-1)*100) {
		t.Fatalf("fx%%=%v", r.FxPct)
	}
	// 1+total = (1+asset)*(blended/spot)
	wantTotal := ((1+0.12)*(0.21/0.204) - 1) * 100
	if !approx(r.TotalReturnPct, wantTotal) {
		t.Fatalf("total%%=%v want %v", r.TotalReturnPct, wantTotal)
	}
	if !r.HomeAvailable {
		t.Fatal("should be available")
	}
	// degraded: no blended
	d := ComputeHolding(2, "AAPL", "USD", 5, 100, 120, "finnhub", 0, 0.204)
	if d.HomeAvailable {
		t.Fatal("no blended -> not available")
	}
	if d.ValueC != 600 {
		t.Fatalf("C side still computed: %+v", d)
	}
}

func TestComputeHoldingNoPrice(t *testing.T) {
	r := ComputeHolding(3, "TSLA", "USD", 5, 100, 0, "finnhub", 0.21, 0.204)
	if r.HomeAvailable {
		t.Fatal("no price -> not available")
	}
	if r.CostC != 0 || r.ValueC != 0 {
		t.Fatalf("no price -> C side zero: %+v", r)
	}
	if r.AssetPnlPct != 0 {
		t.Fatalf("no price -> asset pnl zero: %+v", r)
	}
}

func TestAggregate(t *testing.T) {
	available := ComputeHolding(1, "VOO", "USD", 10, 500, 560, "finnhub", 0.21, 0.204)
	unavailable := ComputeHolding(2, "AAPL", "USD", 5, 100, 120, "finnhub", 0, 0.204)
	other := ComputeHolding(4, "MSFT", "USD", 2, 300, 320, "finnhub", 0.22, 0.21)

	totals := Aggregate([]HoldingResult{available, unavailable, other})

	wantHomeCost := available.HomeCost + other.HomeCost
	wantHomeValue := available.HomeValue + other.HomeValue
	wantTotalReturn := (wantHomeValue - wantHomeCost) / wantHomeCost * 100

	if !approx(totals.HomeCost, wantHomeCost) {
		t.Fatalf("HomeCost=%v want %v", totals.HomeCost, wantHomeCost)
	}
	if !approx(totals.HomeValue, wantHomeValue) {
		t.Fatalf("HomeValue=%v want %v", totals.HomeValue, wantHomeValue)
	}
	if !approx(totals.TotalReturnPct, wantTotalReturn) {
		t.Fatalf("TotalReturnPct=%v want %v", totals.TotalReturnPct, wantTotalReturn)
	}
	if totals.Counted != 2 {
		t.Fatalf("Counted=%v want 2", totals.Counted)
	}
}
