package fx

import "testing"

func TestIsSupported(t *testing.T) {
	for _, c := range []string{"USD", "MYR", "EUR", "JPY"} {
		if !IsSupported(c) {
			t.Fatalf("%s should be supported", c)
		}
	}
	for _, c := range []string{"", "usd", "XXX", "BTC"} {
		if IsSupported(c) {
			t.Fatalf("%s should NOT be supported", c)
		}
	}
	if len(Supported) < 25 {
		t.Fatalf("expected ~30 currencies, got %d", len(Supported))
	}
}
