package fx

import "math"

// Percentile ranks current against window (MyrUsd values, higher = better for
// the buyer). pct is round(100 * fraction of window strictly less than current):
// "today beats pct% of the days in the window". Empty window yields "unknown".
func Percentile(current float64, window []float64) (pct int, assessment string, min, max float64) {
	if len(window) == 0 {
		return 0, "unknown", 0, 0
	}
	min, max = window[0], window[0]
	below := 0
	for _, v := range window {
		if v < current {
			below++
		}
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}
	pct = int(math.Round(100 * float64(below) / float64(len(window))))
	switch {
	case pct >= 67:
		assessment = "good"
	case pct <= 33:
		assessment = "poor"
	default:
		assessment = "middling"
	}
	return pct, assessment, min, max
}
