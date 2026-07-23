// Package price fetches live share prices from Finnhub's free quote endpoint.
package price

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// FinnhubBase is the production Finnhub quote-lookup endpoint. Tests pass an
// httptest server URL instead so Fetch never hits the network in CI.
const FinnhubBase = "https://finnhub.io/api/v1"

// Quote is a single share-price lookup result. Found is false (with no
// error) when Finnhub recognizes the request but has no data for the symbol.
type Quote struct {
	Ticker string
	Price  float64
	Found  bool
}

// quoteResponse mirrors Finnhub's /quote JSON shape. Only the current price
// (c) is used; the rest exists purely for documentation of the payload.
type quoteResponse struct {
	C float64 `json:"c"`
}

// Fetch retrieves the latest price for ticker from base, a Finnhub-compatible
// quote endpoint, authenticated with token. Ticker is used verbatim as the
// Finnhub symbol (uppercased) — US tickers are the plain symbol (VOO, AAPL);
// intl tickers use Finnhub's own symbol format (e.g. 0700.HK).
func Fetch(ctx context.Context, client *http.Client, base, token, ticker string) (Quote, error) {
	sym := strings.ToUpper(ticker)
	reqURL := fmt.Sprintf("%s/quote?symbol=%s&token=%s", base, url.QueryEscape(sym), url.QueryEscape(token))

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
		return Quote{}, fmt.Errorf("finnhub: status %d", resp.StatusCode)
	}

	var body quoteResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return Quote{}, fmt.Errorf("finnhub: malformed response: %w", err)
	}

	if body.C <= 0 {
		return Quote{Ticker: ticker, Found: false}, nil
	}
	return Quote{Ticker: ticker, Price: body.C, Found: true}, nil
}
