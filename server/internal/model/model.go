package model

import "time"

// Conversion records a home-currency → target-currency exchange. Rate is
// target units per 1 home unit, e.g. FromCurrency=MYR, ToCurrency=USD,
// Rate=0.22 means 1 MYR bought 0.22 USD.
type Conversion struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Date         time.Time `json:"date"`
	FromCurrency string    `json:"from_currency"`
	ToCurrency   string    `json:"to_currency"`
	FromAmount   float64   `json:"from_amount"`
	Rate         float64   `json:"rate"`
	Note         string    `json:"note"`
	CreatedAt    time.Time `json:"created_at"`
}

type Holding struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Ticker      string    `json:"ticker"`
	Shares      float64   `json:"shares"`
	AvgCost     float64   `json:"avg_cost"`
	Currency    string    `json:"currency"`
	ManualPrice *float64  `json:"manual_price,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

// Setting is a single-row table (fixed ID 1) holding app-wide preferences,
// currently just the user's home currency.
type Setting struct {
	ID           uint   `gorm:"primaryKey" json:"id"`
	HomeCurrency string `json:"home_currency"`
}
