package fx

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Rate struct {
	UsdMyr    float64   `json:"usd_myr"`
	MyrUsd    float64   `json:"myr_usd"`
	FetchedAt time.Time `json:"fetched_at"`
	Stale     bool      `json:"stale"`
}

// frankfurterResp mirrors https://api.frankfurter.app/latest?from=USD&to=MYR
type frankfurterResp struct {
	Rates map[string]float64 `json:"rates"`
}

func Fetch(ctx context.Context, client *http.Client, baseURL string) (Rate, error) {
	url := baseURL + "/latest?from=USD&to=MYR"
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

	var parsed frankfurterResp
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return Rate{}, err
	}
	usdMyr, ok := parsed.Rates["MYR"]
	if !ok || usdMyr <= 0 {
		return Rate{}, fmt.Errorf("fx upstream: missing or invalid MYR rate")
	}

	return Rate{
		UsdMyr:    usdMyr,
		MyrUsd:    1.0 / usdMyr,
		FetchedAt: time.Now(),
	}, nil
}
