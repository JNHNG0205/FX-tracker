package repository

import (
	"context"
	"testing"
)

func TestSettingsGetDefault(t *testing.T) {
	repo := NewSettingsRepository(testDB(t))
	ctx := context.Background()

	s, err := repo.Get(ctx)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if s.HomeCurrency != "MYR" {
		t.Fatalf("default HomeCurrency = %q, want MYR", s.HomeCurrency)
	}
}

func TestSettingsSetThenGet(t *testing.T) {
	repo := NewSettingsRepository(testDB(t))
	ctx := context.Background()

	if err := repo.SetHomeCurrency(ctx, "EUR"); err != nil {
		t.Fatalf("set: %v", err)
	}
	s, err := repo.Get(ctx)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if s.HomeCurrency != "EUR" {
		t.Fatalf("HomeCurrency = %q, want EUR", s.HomeCurrency)
	}
}

func TestSettingsSetUpsertsSingleRow(t *testing.T) {
	db := testDB(t)
	repo := NewSettingsRepository(db)
	ctx := context.Background()

	if err := repo.SetHomeCurrency(ctx, "EUR"); err != nil {
		t.Fatalf("set EUR: %v", err)
	}
	if err := repo.SetHomeCurrency(ctx, "SGD"); err != nil {
		t.Fatalf("set SGD: %v", err)
	}

	s, err := repo.Get(ctx)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if s.HomeCurrency != "SGD" {
		t.Fatalf("HomeCurrency = %q, want SGD", s.HomeCurrency)
	}

	var count int64
	if err := db.Table("settings").Count(&count).Error; err != nil {
		t.Fatalf("count: %v", err)
	}
	if count != 1 {
		t.Fatalf("settings row count = %d, want 1", count)
	}
}
