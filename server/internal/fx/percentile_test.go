package fx

import "testing"

func TestPercentile(t *testing.T) {
	tests := []struct {
		name     string
		current  float64
		window   []float64
		wantPct  int
		wantVerd string
		wantMin  float64
		wantMax  float64
	}{
		{"above all is good", 5, []float64{1, 2, 3, 4}, 100, "good", 1, 4},
		{"below all is poor", 0, []float64{1, 2, 3, 4}, 0, "poor", 1, 4},
		{"middle is middling", 2.5, []float64{1, 2, 3, 4}, 50, "middling", 1, 4},
		{"strict-less at a member", 3, []float64{1, 2, 3, 4}, 50, "middling", 1, 4},
		{"67th percentile is good", 67, hundred(), 67, "good", 0, 99},
		{"66th percentile is middling", 66, hundred(), 66, "middling", 0, 99},
		{"33rd percentile is poor", 33, hundred(), 33, "poor", 0, 99},
		{"34th percentile is middling", 34, hundred(), 34, "middling", 0, 99},
		{"empty window is unknown", 4.1, nil, 0, "unknown", 0, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pct, verd, min, max := Percentile(tt.current, tt.window)
			if pct != tt.wantPct || verd != tt.wantVerd || min != tt.wantMin || max != tt.wantMax {
				t.Fatalf("Percentile(%v,…) = (%d,%q,%v,%v), want (%d,%q,%v,%v)",
					tt.current, pct, verd, min, max, tt.wantPct, tt.wantVerd, tt.wantMin, tt.wantMax)
			}
		})
	}
}

// hundred returns 0..99.
func hundred() []float64 {
	s := make([]float64, 100)
	for i := range s {
		s[i] = float64(i)
	}
	return s
}
