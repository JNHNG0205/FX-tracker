// Package price fetches live share prices from Stooq's free CSV endpoint.
package price

import (
	"context"
	"encoding/csv"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// StooqBase is the production Stooq quote-lookup endpoint. Tests pass an
// httptest server URL instead so Fetch never hits the network in CI.
const StooqBase = "https://stooq.com/q/l/"

// Quote is a single share-price lookup result. Found is false (with no
// error) when Stooq recognizes the request but has no data for the symbol.
type Quote struct {
	Ticker string
	Price  float64
	Found  bool
}

// Fetch retrieves the latest close price for ticker/currency from base, a
// Stooq-compatible CSV quote endpoint.
func Fetch(ctx context.Context, client *http.Client, base, ticker, currency string) (Quote, error) {
	sym := strings.ToLower(ticker)
	if currency == "USD" {
		sym += ".us"
	}
	reqURL := fmt.Sprintf("%s?s=%s&f=sd2t2ohlcv&h&e=csv", base, url.QueryEscape(sym))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return Quote{}, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return Quote{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Quote{}, fmt.Errorf("stooq: status %d", resp.StatusCode)
	}

	rows, err := csv.NewReader(resp.Body).ReadAll()
	if err != nil || len(rows) < 2 || len(rows[1]) < 7 {
		return Quote{}, fmt.Errorf("stooq: malformed response")
	}

	close := rows[1][6]
	if close == "N/D" || close == "" {
		return Quote{Ticker: ticker, Found: false}, nil
	}
	p, err := strconv.ParseFloat(close, 64)
	if err != nil || p <= 0 {
		return Quote{Ticker: ticker, Found: false}, nil
	}
	return Quote{Ticker: ticker, Price: p, Found: true}, nil
}
