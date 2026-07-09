package fx

import "testing"

func TestAssess(t *testing.T) {
	tests := []struct {
		name              string
		current, min, max float64
		want              string
	}{
		{"top third is good", 0.2130, 0.2100, 0.2130, "good"},
		{"bottom third is poor", 0.2100, 0.2100, 0.2130, "poor"},
		{"middle is middling", 0.2115, 0.2100, 0.2130, "middling"},
		{"no range yet is unknown", 0.2115, 0.2115, 0.2115, "unknown"},
		{"zero max is unknown", 0.0, 0.0, 0.0, "unknown"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Assess(tt.current, tt.min, tt.max); got != tt.want {
				t.Fatalf("Assess(%v,%v,%v) = %q, want %q", tt.current, tt.min, tt.max, got, tt.want)
			}
		})
	}
}
