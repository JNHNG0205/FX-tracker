package fx

// Supported is the ECB/frankfurter fiat set — the source of truth for
// currency validation and the /api/currencies response.
var Supported = map[string]string{
	"AUD": "Australian Dollar", "BRL": "Brazilian Real", "CAD": "Canadian Dollar",
	"CHF": "Swiss Franc", "CNY": "Chinese Renminbi Yuan", "CZK": "Czech Koruna",
	"DKK": "Danish Krone", "EUR": "Euro", "GBP": "British Pound",
	"HKD": "Hong Kong Dollar", "HUF": "Hungarian Forint", "IDR": "Indonesian Rupiah",
	"ILS": "Israeli New Shekel", "INR": "Indian Rupee", "ISK": "Icelandic Króna",
	"JPY": "Japanese Yen", "KRW": "South Korean Won", "MXN": "Mexican Peso",
	"MYR": "Malaysian Ringgit", "NOK": "Norwegian Krone", "NZD": "New Zealand Dollar",
	"PHP": "Philippine Peso", "PLN": "Polish Złoty", "RON": "Romanian Leu",
	"SEK": "Swedish Krona", "SGD": "Singapore Dollar", "THB": "Thai Baht",
	"TRY": "Turkish Lira", "USD": "United States Dollar", "ZAR": "South African Rand",
}

type Currency struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

func IsSupported(code string) bool {
	_, ok := Supported[code]
	return ok
}
