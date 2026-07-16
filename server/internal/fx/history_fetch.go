package fx

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"time"
)

type HistPoint struct {
	Date  time.Time
	Value float64
}

// histResp mirrors frankfurter's time-series shape: rates keyed by date, each a
// currency->value map. e.g. {"rates":{"2026-06-09":{"USD":0.245}}}.
type histResp struct {
	Rates map[string]map[string]float64 `json:"rates"`
}

// FetchHistory fetches the from/to daily series in [start,end]. Business-day
// gaps are simply absent points. The client follows the 301 that frankfurter's
// historical URLs issue (default http.Client behavior).
func FetchHistory(ctx context.Context, client *http.Client, baseURL, from, to string, start, end time.Time) ([]HistPoint, error) {
	url := fmt.Sprintf("%s/%s..%s?from=%s&to=%s",
		baseURL, start.Format("2006-01-02"), end.Format("2006-01-02"), from, to)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fx history: status %d", resp.StatusCode)
	}

	var parsed histResp
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return nil, err
	}

	pts := make([]HistPoint, 0, len(parsed.Rates))
	for dateStr, rates := range parsed.Rates {
		v, ok := rates[to]
		if !ok || v <= 0 {
			continue
		}
		d, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			continue
		}
		pts = append(pts, HistPoint{Date: d, Value: v})
	}
	sort.Slice(pts, func(i, j int) bool { return pts[i].Date.Before(pts[j].Date) })
	return pts, nil
}
