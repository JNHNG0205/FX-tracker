package model

import "time"

type Conversion struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	Date       time.Time `json:"date"`
	MyrAmount  float64   `json:"myr_amount"`
	RateMyrUsd float64   `json:"rate_myr_usd"`
	Note       string    `json:"note"`
	CreatedAt  time.Time `json:"created_at"`
}

type Holding struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	Ticker     string    `json:"ticker"`
	Shares     float64   `json:"shares"`
	AvgCostUsd float64   `json:"avg_cost_usd"`
	CreatedAt  time.Time `json:"created_at"`
}
