package repository

import (
	"context"
	"os"
	"testing"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"fx-tracker/internal/model"
)

func testDB(t *testing.T) *gorm.DB {
	t.Helper()
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		url = "postgres://fx:fx@localhost:5432/fxtracker_test?sslmode=disable"
	}
	db, err := gorm.Open(postgres.Open(url), &gorm.Config{})
	if err != nil {
		t.Fatalf("connect test db (is Postgres up and fxtracker_test created?): %v", err)
	}
	if err := db.AutoMigrate(&model.Conversion{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	if err := db.Exec("TRUNCATE conversions RESTART IDENTITY").Error; err != nil {
		t.Fatalf("truncate: %v", err)
	}
	return db
}

func TestCreateAndList(t *testing.T) {
	repo := NewConversionRepository(testDB(t))
	ctx := context.Background()
	older := &model.Conversion{Date: time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC), MyrAmount: 1000, RateMyrUsd: 0.22}
	newer := &model.Conversion{Date: time.Date(2026, 7, 8, 0, 0, 0, 0, time.UTC), MyrAmount: 500, RateMyrUsd: 0.24}
	if err := repo.Create(ctx, older); err != nil {
		t.Fatal(err)
	}
	if err := repo.Create(ctx, newer); err != nil {
		t.Fatal(err)
	}
	list, err := repo.List(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 2 || !list[0].Date.Equal(newer.Date) {
		t.Fatalf("expected newest first; got %+v", list)
	}
}

func TestBlendedRate(t *testing.T) {
	repo := NewConversionRepository(testDB(t))
	ctx := context.Background()

	// Empty → zeros, no error.
	rate, tm, tu, err := repo.BlendedRate(ctx)
	if err != nil || rate != 0 || tm != 0 || tu != 0 {
		t.Fatalf("empty blended = (%v,%v,%v,%v)", rate, tm, tu, err)
	}

	// 1000 MYR @0.20 → 200 USD; 1000 MYR @0.30 → 300 USD. Blended = 500/2000 = 0.25.
	_ = repo.Create(ctx, &model.Conversion{Date: time.Now(), MyrAmount: 1000, RateMyrUsd: 0.20})
	_ = repo.Create(ctx, &model.Conversion{Date: time.Now(), MyrAmount: 1000, RateMyrUsd: 0.30})
	rate, tm, tu, err = repo.BlendedRate(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if rate != 0.25 || tm != 2000 || tu != 500 {
		t.Fatalf("blended = (rate %v, myr %v, usd %v), want (0.25, 2000, 500)", rate, tm, tu)
	}
}
