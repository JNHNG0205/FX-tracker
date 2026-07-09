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

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(&model.Conversion{}, &model.Holding{})
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
