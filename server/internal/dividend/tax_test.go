package dividend

import (
	"testing"

	"github.com/shopspring/decimal"
)

func d(s string) decimal.Decimal { return decimal.RequireFromString(s) }

func TestWithhold(t *testing.T) {
	tests := []struct{ amount, currency, wantW, wantN string }{
		{"100", "USD", "30.00", "70.00"},
		{"10.01", "USD", "3.00", "7.01"},   // 3.003 rounds to 3.00
		{"33.33", "USD", "10.00", "23.33"}, // 9.999 rounds to 10.00
		{"100", "EUR", "0", "100"},         // non-USD: no withholding
		{"250.55", "MYR", "0", "250.55"},
	}
	for _, tt := range tests {
		w, n := Withhold(d(tt.amount), tt.currency)
		if !w.Equal(d(tt.wantW)) || !n.Equal(d(tt.wantN)) {
			t.Fatalf("Withhold(%s,%s) = (%s,%s), want (%s,%s)", tt.amount, tt.currency, w, n, tt.wantW, tt.wantN)
		}
	}
}
