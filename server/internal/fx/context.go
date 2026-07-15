package fx

import "time"

// Assessment ranks a rate against a historical window ("7d", "30d", etc).
type Assessment struct {
	Label      string  `json:"label"`
	Assessment string  `json:"assessment"`
	Percentile int     `json:"percentile"`
	Min        float64 `json:"min"`
	Max        float64 `json:"max"`
	Samples    int     `json:"samples"`
	Start      string  `json:"start"`
	End        string  `json:"end"`
}

// timeframe: Days==0 means year-to-date.
type timeframe struct {
	label string
	days  int
}

var timeframes = []timeframe{
	{"7d", 7}, {"14d", 14}, {"30d", 30}, {"90d", 90}, {"YTD", 0},
}

// windowStart returns the earliest date included for a timeframe as of `asOf`.
func windowStart(tf timeframe, asOf time.Time) time.Time {
	if tf.days == 0 { // YTD
		return time.Date(asOf.Year(), time.January, 1, 0, 0, 0, 0, asOf.Location())
	}
	return asOf.AddDate(0, 0, -tf.days)
}

// assess ranks current against each timeframe's window, sliced from points
// (ascending by date).
func assess(current float64, points []HistPoint, asOf time.Time) []Assessment {
	out := make([]Assessment, 0, len(timeframes))
	for _, tf := range timeframes {
		start := windowStart(tf, asOf)
		var window []float64
		var firstDate, lastDate time.Time
		for _, p := range points {
			if p.Date.Before(start) {
				continue
			}
			if window == nil {
				firstDate = p.Date
			}
			lastDate = p.Date
			window = append(window, p.Value)
		}
		pct, verdict, min, max := Percentile(current, window)
		a := Assessment{
			Label:      tf.label,
			Assessment: verdict,
			Percentile: pct,
			Min:        min,
			Max:        max,
			Samples:    len(window),
		}
		if len(window) > 0 {
			a.Start = firstDate.Format("2006-01-02")
			a.End = lastDate.Format("2006-01-02")
		}
		out = append(out, a)
	}
	return out
}

// Point is one day of a pair's series in the /api/rate/history payload.
type Point struct {
	Date  string  `json:"date"`
	Value float64 `json:"value"`
}

// RateHistory is the /api/rate/history payload for a currency pair.
type RateHistory struct {
	From   string  `json:"from"`
	To     string  `json:"to"`
	Points []Point `json:"points"`
	Stale  bool    `json:"stale"`
}

// RateContext is the /api/rate/context payload: the live rate for a pair
// plus a per-timeframe historical assessment.
type RateContext struct {
	From         string       `json:"from"`
	To           string       `json:"to"`
	Rate         float64      `json:"rate"`
	Inverse      float64      `json:"inverse"`
	Stale        bool         `json:"stale"`
	FetchedAt    time.Time    `json:"fetched_at"`
	HistoryStale bool         `json:"history_stale"`
	Timeframes   []Assessment `json:"timeframes"`
}
