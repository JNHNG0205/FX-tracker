package repository

import (
	"context"
	"errors"
	"testing"

	"gorm.io/gorm"

	"fx-tracker/internal/model"
)

func TestHoldingCRUD(t *testing.T) {
	repo := NewHoldingRepository(testDB(t))
	ctx := context.Background()
	h := &model.Holding{Ticker: "VOO", Shares: 10, AvgCost: 500, Currency: "USD"}
	if err := repo.Create(ctx, h); err != nil {
		t.Fatal(err)
	}
	list, _ := repo.List(ctx)
	if len(list) != 1 || list[0].Ticker != "VOO" {
		t.Fatalf("list = %+v", list)
	}
	h.Shares = 12
	if err := repo.Update(ctx, h); err != nil {
		t.Fatal(err)
	}
	list, _ = repo.List(ctx)
	if list[0].Shares != 12 {
		t.Fatalf("update not applied: %+v", list[0])
	}
	if err := repo.Delete(ctx, h.ID); err != nil {
		t.Fatal(err)
	}
	if err := repo.Delete(ctx, h.ID); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("delete missing: %v", err)
	}
}

func TestHoldingUpdateManualPrice(t *testing.T) {
	repo := NewHoldingRepository(testDB(t))
	ctx := context.Background()
	h := &model.Holding{Ticker: "VOO", Shares: 10, AvgCost: 500, Currency: "USD"}
	if err := repo.Create(ctx, h); err != nil {
		t.Fatal(err)
	}

	price := 550.5
	h.ManualPrice = &price
	if err := repo.Update(ctx, h); err != nil {
		t.Fatalf("update set manual_price: %v", err)
	}
	list, _ := repo.List(ctx)
	if list[0].ManualPrice == nil || *list[0].ManualPrice != price {
		t.Fatalf("manual_price not set: %+v", list[0])
	}

	h.ManualPrice = nil
	if err := repo.Update(ctx, h); err != nil {
		t.Fatalf("update clear manual_price: %v", err)
	}
	list, _ = repo.List(ctx)
	if list[0].ManualPrice != nil {
		t.Fatalf("manual_price not cleared: %+v", list[0])
	}
}
