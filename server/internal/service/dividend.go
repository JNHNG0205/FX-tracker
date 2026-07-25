package service

import (
	"context"

	"github.com/shopspring/decimal"

	"fx-tracker/internal/dividend"
	"fx-tracker/internal/model"
	"fx-tracker/internal/repository"
)

// DividendEntry is the wire representation of a single dividend, with money
// fields carried as decimal strings to avoid float precision loss in transit.
type DividendEntry struct {
	ID            uint    `json:"id"`
	Ticker        string  `json:"ticker"`
	Currency      string  `json:"currency"`
	Date          string  `json:"date"`
	Note          string  `json:"note"`
	Amount        string  `json:"amount"`
	Withholding   string  `json:"withholding"`
	Net           string  `json:"net"`
	HomeNet       float64 `json:"home_net"`
	HomeAvailable bool    `json:"home_available"`
}

// CurrencyTotal aggregates gross/withholding/net across all dividends in a
// single currency, still in that currency's decimal units.
type CurrencyTotal struct {
	Currency    string `json:"currency"`
	Gross       string `json:"gross"`
	Withholding string `json:"withholding"`
	Net         string `json:"net"`
}

// DividendTotals rolls up per-currency totals plus the home-currency net
// (spot-approximated, summed only over entries where a spot rate was
// available).
type DividendTotals struct {
	ByCurrency []CurrencyTotal `json:"by_currency"`
	HomeNet    float64         `json:"home_net"`
}

// DividendSummary is the full response for the dividend summary endpoint.
type DividendSummary struct {
	HomeCurrency string          `json:"home_currency"`
	Dividends    []DividendEntry `json:"dividends"`
	Totals       DividendTotals  `json:"totals"`
}

// currencyAccumulator holds running decimal totals for one currency while we
// walk the dividend list in first-seen order.
type currencyAccumulator struct {
	currency    string
	gross       decimal.Decimal
	withholding decimal.Decimal
	net         decimal.Decimal
}

// DividendService delegates CRUD to the repository and computes the
// withholding/net summary, approximated into home currency at spot.
type DividendService struct {
	repo repository.DividendRepository
	spot SpotSource
}

func NewDividendService(repo repository.DividendRepository, spot SpotSource) *DividendService {
	return &DividendService{repo: repo, spot: spot}
}

func (s *DividendService) Create(ctx context.Context, d *model.Dividend) error {
	return s.repo.Create(ctx, d)
}

func (s *DividendService) List(ctx context.Context) ([]model.Dividend, error) {
	return s.repo.List(ctx)
}

func (s *DividendService) Update(ctx context.Context, d *model.Dividend) error {
	return s.repo.Update(ctx, d)
}

func (s *DividendService) Delete(ctx context.Context, id uint) error {
	return s.repo.Delete(ctx, id)
}

// Summary computes per-dividend withholding/net (exact decimal) plus a
// home-currency approximation at spot, and per-currency + combined totals.
// A per-dividend spot lookup failure degrades that entry (HomeAvailable
// false) rather than failing the whole response, mirroring PortfolioService.
func (s *DividendService) Summary(ctx context.Context, home string) (DividendSummary, error) {
	dividends, err := s.repo.List(ctx)
	if err != nil {
		return DividendSummary{}, err
	}

	entries := make([]DividendEntry, len(dividends))
	var order []string
	accs := make(map[string]*currencyAccumulator)
	var totalHomeNet float64

	for i, d := range dividends {
		w, n := dividend.Withhold(d.Amount, d.Currency)

		var spot float64
		var available bool
		if home == d.Currency {
			spot, available = 1, true
		} else if r, err := s.spot.Rate(ctx, home, d.Currency); err == nil && r.Rate > 0 {
			spot, available = r.Rate, true
		}

		var homeNet float64
		if available {
			homeNet = n.InexactFloat64() / spot
			totalHomeNet += homeNet
		}

		entries[i] = DividendEntry{
			ID:            d.ID,
			Ticker:        d.Ticker,
			Currency:      d.Currency,
			Date:          d.Date.Format("2006-01-02"),
			Note:          d.Note,
			Amount:        d.Amount.StringFixed(2),
			Withholding:   w.StringFixed(2),
			Net:           n.StringFixed(2),
			HomeNet:       homeNet,
			HomeAvailable: available,
		}

		acc, ok := accs[d.Currency]
		if !ok {
			acc = &currencyAccumulator{currency: d.Currency}
			accs[d.Currency] = acc
			order = append(order, d.Currency)
		}
		acc.gross = acc.gross.Add(d.Amount)
		acc.withholding = acc.withholding.Add(w)
		acc.net = acc.net.Add(n)
	}

	byCurrency := make([]CurrencyTotal, len(order))
	for i, cur := range order {
		acc := accs[cur]
		byCurrency[i] = CurrencyTotal{
			Currency:    acc.currency,
			Gross:       acc.gross.StringFixed(2),
			Withholding: acc.withholding.StringFixed(2),
			Net:         acc.net.StringFixed(2),
		}
	}

	return DividendSummary{
		HomeCurrency: home,
		Dividends:    entries,
		Totals: DividendTotals{
			ByCurrency: byCurrency,
			HomeNet:    totalHomeNet,
		},
	}, nil
}
