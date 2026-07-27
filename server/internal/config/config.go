package config

import (
	"bufio"
	"os"
	"strings"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"fx-tracker/internal/model"
)

type Config struct {
	DatabaseURL   string
	Port          string
	FxBaseURL     string
	FinnhubAPIKey string
}

func Load() Config {
	// Load a local .env into the process env if present. Real environment
	// variables always win, so this only fills what isn't already set.
	loadDotEnv(".env")
	return Config{
		DatabaseURL:   env("DATABASE_URL", "postgres://fx:fx@localhost:5432/fxtracker?sslmode=disable"),
		Port:          env("PORT", "8080"),
		FxBaseURL:     env("FX_BASE_URL", "https://api.frankfurter.app"),
		FinnhubAPIKey: env("FINNHUB_API_KEY", ""),
	}
}

// loadDotEnv reads KEY=VALUE lines from path and sets any that aren't already
// present in the environment. A missing file is not an error.
func loadDotEnv(path string) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, val, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		val = strings.Trim(strings.TrimSpace(val), `"'`)
		if key != "" && os.Getenv(key) == "" {
			_ = os.Setenv(key, val)
		}
	}
}

func Connect(cfg Config) (*gorm.DB, error) {
	return gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{})
}

// Migrate applies AutoMigrate and then backfills the Conversion table from
// its pre-multi-currency shape (myr_amount/rate_myr_usd) to the generalized
// from_currency/to_currency/from_amount/rate columns. It is a no-op on a
// fresh DB (no old columns) and idempotent (re-running does nothing, since
// the old columns are dropped after the one-time backfill).
func Migrate(db *gorm.DB) error {
	if err := db.AutoMigrate(&model.Conversion{}, &model.Holding{}, &model.Setting{}, &model.Dividend{}); err != nil {
		return err
	}

	m := db.Migrator()
	if m.HasColumn(&model.Conversion{}, "myr_amount") {
		if err := db.Exec(`UPDATE conversions SET from_currency='MYR', to_currency='USD', from_amount=myr_amount, rate=rate_myr_usd WHERE (from_currency IS NULL OR from_currency='')`).Error; err != nil {
			return err
		}
		_ = m.DropColumn(&model.Conversion{}, "myr_amount")
		_ = m.DropColumn(&model.Conversion{}, "rate_myr_usd")
	}
	if m.HasColumn(&model.Holding{}, "avg_cost_usd") {
		_ = m.DropColumn(&model.Holding{}, "avg_cost_usd")
	}
	return nil
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
