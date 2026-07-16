package config

import (
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"fx-tracker/internal/model"
)

type Config struct {
	DatabaseURL string
	Port        string
	FxBaseURL   string
}

func Load() Config {
	return Config{
		DatabaseURL: env("DATABASE_URL", "postgres://fx:fx@localhost:5432/fxtracker?sslmode=disable"),
		Port:        env("PORT", "8080"),
		FxBaseURL:   env("FX_BASE_URL", "https://api.frankfurter.app"),
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
	if err := db.AutoMigrate(&model.Conversion{}, &model.Holding{}, &model.Setting{}); err != nil {
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
	return nil
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
