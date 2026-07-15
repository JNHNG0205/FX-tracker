package fx

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Rate struct {
	From      string    `json:"from"`
	To        string    `json:"to"`
	Rate      float64   `json:"rate"`    // target per 1 home
	Inverse   float64   `json:"inverse"` // home per 1 target
	FetchedAt time.Time `json:"fetched_at"`
	Stale     bool      `json:"stale"`
}

// latestResp mirrors https://api.frankfurter.app/latest?from=MYR&to=USD
type latestResp struct {
	Rates map[string]float64 `json:"rates"`
}

func Fetch(ctx context.Context, client *http.Client, baseURL, from, to string) (Rate, error) {
	url := fmt.Sprintf("%s/latest?from=%s&to=%s", baseURL, from, to)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return Rate{}, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return Rate{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Rate{}, fmt.Errorf("fx upstream: status %d", resp.StatusCode)
	}

	var parsed latestResp
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return Rate{}, err
	}
	v, ok := parsed.Rates[to]
	if !ok || v <= 0 {
		return Rate{}, fmt.Errorf("fx upstream: missing or invalid rate for %s", to)
	}

	return Rate{
		From:      from,
		To:        to,
		Rate:      v,
		Inverse:   1.0 / v,
		FetchedAt: time.Now(),
	}, nil
}
