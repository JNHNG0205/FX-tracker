package dividend

import "github.com/shopspring/decimal"

// usRate is the US dividend withholding for Malaysian residents (no treaty).
var usRate = decimal.RequireFromString("0.30")

// Withhold returns the tax withheld and the net dividend. US (USD) dividends
// are withheld at 30%, rounded to 2 dp; all other currencies are untaxed here.
func Withhold(amount decimal.Decimal, currency string) (withholding, net decimal.Decimal) {
	rate := decimal.Zero
	if currency == "USD" {
		rate = usRate
	}
	withholding = amount.Mul(rate).Round(2)
	net = amount.Sub(withholding)
	return withholding, net
}
