package config

import (
	"os"
	"testing"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"fx-tracker/internal/model"
)

// testConfigDB connects to a dedicated fxtracker_migtest database rather than
// the shared fxtracker_test database used by repository tests. Migration
// tests here DROP/CREATE the conversions table directly, which would race
// repository tests truncating/querying that same table when packages run
// concurrently under `go test ./... -race`. Isolating the database sidesteps
// the race entirely instead of trying to serialize the two test binaries.
func testConfigDB(t *testing.T) *gorm.DB {
	t.Helper()

	maintURL := "postgres://fx:fx@localhost:5432/fxtracker?sslmode=disable"
	maintDB, err := gorm.Open(postgres.Open(maintURL), &gorm.Config{})
	if err != nil {
		t.Skipf("postgres unreachable, skipping migration test: %v", err)
	}
	// Ignore the error: it's almost always "database already exists"
	// (SQLSTATE 42P04), and CREATE DATABASE has no IF NOT EXISTS form.
	_ = maintDB.Exec("CREATE DATABASE fxtracker_migtest").Error

	url := os.Getenv("MIG_TEST_DATABASE_URL")
	if url == "" {
		url = "postgres://fx:fx@localhost:5432/fxtracker_migtest?sslmode=disable"
	}
	db, err := gorm.Open(postgres.Open(url), &gorm.Config{})
	if err != nil {
		t.Fatalf("connect fxtracker_migtest (is Postgres up?): %v", err)
	}
	return db
}

// TestMigrateBackfillsOldConversions simulates a DB still on the pre-multi-
// currency conversions schema (myr_amount/rate_myr_usd) and asserts that
// Migrate backfills the new currency-pair columns and drops the old ones.
func TestMigrateBackfillsOldConversions(t *testing.T) {
	db := testConfigDB(t)

	if err := db.Exec(`DROP TABLE IF EXISTS conversions`).Error; err != nil {
		t.Fatalf("drop table: %v", err)
	}
	t.Cleanup(func() {
		_ = db.Exec(`DROP TABLE IF EXISTS conversions`).Error
	})

	if err := db.Exec(`CREATE TABLE conversions (
		id serial PRIMARY KEY,
		date timestamptz,
		myr_amount double precision,
		rate_myr_usd double precision,
		note text,
		created_at timestamptz
	)`).Error; err != nil {
		t.Fatalf("create old-shape table: %v", err)
	}

	if err := db.Exec(`INSERT INTO conversions (date, myr_amount, rate_myr_usd, note, created_at)
		VALUES (now(), 1000, 0.22, 'legacy', now())`).Error; err != nil {
		t.Fatalf("insert legacy row: %v", err)
	}

	if err := Migrate(db); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	var row struct {
		FromCurrency string
		ToCurrency   string
		FromAmount   float64
		Rate         float64
	}
	if err := db.Raw(`SELECT from_currency, to_currency, from_amount, rate FROM conversions LIMIT 1`).Scan(&row).Error; err != nil {
		t.Fatalf("select backfilled row: %v", err)
	}
	if row.FromCurrency != "MYR" || row.ToCurrency != "USD" || row.FromAmount != 1000 || row.Rate != 0.22 {
		t.Fatalf("backfilled row = %+v, want {MYR USD 1000 0.22}", row)
	}

	m := db.Migrator()
	if m.HasColumn(&model.Conversion{}, "myr_amount") {
		t.Fatalf("myr_amount column still present after migration")
	}
	if m.HasColumn(&model.Conversion{}, "rate_myr_usd") {
		t.Fatalf("rate_myr_usd column still present after migration")
	}

	// Re-running Migrate must be idempotent: no old columns to backfill from,
	// so the row (and column set) should be unchanged.
	if err := Migrate(db); err != nil {
		t.Fatalf("second migrate: %v", err)
	}
	var row2 struct {
		FromCurrency string
		ToCurrency   string
		FromAmount   float64
		Rate         float64
	}
	if err := db.Raw(`SELECT from_currency, to_currency, from_amount, rate FROM conversions LIMIT 1`).Scan(&row2).Error; err != nil {
		t.Fatalf("select after second migrate: %v", err)
	}
	if row2 != row {
		t.Fatalf("row changed after idempotent re-migrate: got %+v, want %+v", row2, row)
	}
}

// TestMigrateFreshDBNoOp verifies the backfill branch is skipped entirely
// when there is no legacy myr_amount column (a brand-new database).
func TestMigrateFreshDBNoOp(t *testing.T) {
	db := testConfigDB(t)

	if err := db.Exec(`DROP TABLE IF EXISTS conversions`).Error; err != nil {
		t.Fatalf("drop table: %v", err)
	}
	t.Cleanup(func() {
		_ = db.Exec(`DROP TABLE IF EXISTS conversions`).Error
	})

	if err := Migrate(db); err != nil {
		t.Fatalf("migrate fresh db: %v", err)
	}

	m := db.Migrator()
	if !m.HasTable(&model.Conversion{}) {
		t.Fatalf("conversions table not created")
	}
	if m.HasColumn(&model.Conversion{}, "myr_amount") {
		t.Fatalf("myr_amount column unexpectedly present on fresh db")
	}
}
