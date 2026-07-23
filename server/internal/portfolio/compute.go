// Package portfolio computes MYR true-return for a holding by splitting
// total return into an asset-price component and an FX component. All
// functions here are pure (no IO) so the math is exhaustively testable
// without a DB or network.
package portfolio

// HoldingResult is the per-holding return breakdown, ready to serialize.
type HoldingResult struct {
	ID          uint    `json:"id"`
	Ticker      string  `json:"ticker"`
	Currency    string  `json:"currency"`
	Shares      float64 `json:"shares"`
	AvgCost     float64 `json:"avg_cost"`
	Price       float64 `json:"price"`
	PriceSource string  `json:"price_source"`

	CostC       float64 `json:"cost_c"`
	ValueC      float64 `json:"value_c"`
	AssetPnlPct float64 `json:"asset_pnl_pct"`

	Blended float64 `json:"blended"`
	Spot    float64 `json:"spot"`

	HomeCost       float64 `json:"home_cost"`
	HomeValue      float64 `json:"home_value"`
	TotalReturnPct float64 `json:"total_return_pct"`
	FxPct          float64 `json:"fx_pct"`

	HomeAvailable bool `json:"home_available"`
}

// Totals aggregates the home-currency return across all holdings for which
// a home-currency conversion was available.
type Totals struct {
	HomeCost       float64 `json:"home_cost"`
	HomeValue      float64 `json:"home_value"`
	TotalReturnPct float64 `json:"total_return_pct"`
	Counted        int     `json:"counted"`
}

// Response is the top-level portfolio return payload.
type Response struct {
	HomeCurrency string          `json:"home_currency"`
	Holdings     []HoldingResult `json:"holdings"`
	Totals       Totals          `json:"totals"`
}

// ComputeHolding builds the return breakdown for one holding.
//
// blended and spot are both expressed as target-per-1-home (e.g. USD per 1
// MYR): blended is the cost-weighted rate paid when converting home currency
// into the asset's currency, spot is the current rate. Dividing a
// target-currency amount by one of these rates converts it back to home
// currency.
func ComputeHolding(id uint, ticker, currency string, shares, avgCost, price float64, priceSource string, blended, spot float64) HoldingResult {
	r := HoldingResult{
		ID:          id,
		Ticker:      ticker,
		Currency:    currency,
		Shares:      shares,
		AvgCost:     avgCost,
		Price:       price,
		PriceSource: priceSource,
		Blended:     blended,
		Spot:        spot,
	}

	if price > 0 {
		r.CostC = shares * avgCost
		r.ValueC = shares * price
		if r.CostC > 0 {
			r.AssetPnlPct = (r.ValueC - r.CostC) / r.CostC * 100
		}
	}

	r.HomeAvailable = price > 0 && blended > 0 && spot > 0
	if !r.HomeAvailable {
		return r
	}

	r.HomeCost = r.CostC / blended
	r.HomeValue = r.ValueC / spot
	if r.HomeCost > 0 {
		r.TotalReturnPct = (r.HomeValue - r.HomeCost) / r.HomeCost * 100
	}
	r.FxPct = (blended/spot - 1) * 100

	return r
}

// Aggregate sums the home-currency cost/value across all holdings that have
// a home-currency conversion available, and derives a blended total return.
func Aggregate(rs []HoldingResult) Totals {
	var t Totals
	for _, r := range rs {
		if !r.HomeAvailable {
			continue
		}
		t.HomeCost += r.HomeCost
		t.HomeValue += r.HomeValue
		t.Counted++
	}
	if t.HomeCost > 0 {
		t.TotalReturnPct = (t.HomeValue - t.HomeCost) / t.HomeCost * 100
	}
	return t
}
