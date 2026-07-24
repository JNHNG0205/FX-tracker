package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	"fx-tracker/internal/model"
)

func TestDividendCRUD(t *testing.T) {
	repo := NewDividendRepository(testDB(t))
	ctx := context.Background()
	d := &model.Dividend{Ticker: "VOO", Currency: "USD", Amount: decimal.RequireFromString("12.34"), Date: time.Now()}
	if err := repo.Create(ctx, d); err != nil {
		t.Fatal(err)
	}
	list, _ := repo.List(ctx)
	if len(list) != 1 || !list[0].Amount.Equal(decimal.RequireFromString("12.34")) {
		t.Fatalf("amount didn't round-trip: %+v", list)
	}
	d.Amount = decimal.RequireFromString("56.78")
	if err := repo.Update(ctx, d); err != nil {
		t.Fatal(err)
	}
	list, _ = repo.List(ctx)
	if !list[0].Amount.Equal(decimal.RequireFromString("56.78")) {
		t.Fatalf("update: %+v", list[0])
	}
	if err := repo.Delete(ctx, d.ID); err != nil {
		t.Fatal(err)
	}
	if err := repo.Delete(ctx, d.ID); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("delete missing: %v", err)
	}
}
