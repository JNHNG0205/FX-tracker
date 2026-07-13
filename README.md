# MYR → USD Tool

A personal finance tool for a Malaysia-based investor who buys US assets (VOO, US stocks) through Moomoo, funding in MYR. It covers the layer broker apps ignore: **MYR + FX + tax**.

## Status

Under active development. Building **Phase 0–1** first: a live MYR/USD FX checker, end-to-end.

## Roadmap

- **Phase 0–1 — FX checker.** Live mid-market MYR→USD (and USD→MYR), with a good/middling/poor assessment based on the recent range. Auto-refreshing, with graceful degradation to the last-known rate.
- **Phase 2 — Conversion log + DCA planner.** Log MYR→USD conversions, compute a blended average rate, and compare today's rate against it.
- **Phase 3 — MYR true-return tracker.** Manual holdings, live USD prices, and real return translated to MYR using your blended conversion rate — not spot.
- **Phase 4 — Dividend + withholding tracker.** Projected dividends with the 30% US withholding tax applied for Malaysian residents.

## Tech Stack

- **Backend:** Go 1.22+, Gin, GORM, PostgreSQL
- **Frontend:** React + Vite + TypeScript (strict), TanStack Query, Tailwind v4, shadcn/ui
- **Package manager (frontend):** Bun
- **Local database:** Docker Compose Postgres

## Getting started

Requires Docker, Go 1.22+, and Bun.

```bash
# 1. Start Postgres (user/pass/db: fx/fx/fxtracker on :5432)
docker compose up -d

# 2. Backend — API on :8080
cd server
go mod tidy
go run ./cmd/server

# 3. Frontend — UI on :5173 (proxies /api to :8080)
cd web
bun install
bun run dev
```

## API

| Endpoint | Returns |
| --- | --- |
| `GET /health` | `{"status":"ok"}` |
| `GET /api/rate` | `{ usd_myr, myr_usd, fetched_at, stale }` — mid-market USD/MYR |
| `GET /api/rate/context` | live rate + a `timeframes` array (7d/14d/30d/90d/YTD), each with a percentile-based `assessment` |
| `POST /api/conversions` | log a conversion `{ date?, myr_amount, rate_myr_usd, note? }` (amount & rate must be > 0) |
| `GET /api/conversions` | conversions, newest first |
| `GET /api/conversions/status` | blended average rate + DCA compare vs the live rate |
| `GET /api/rate/history` | `{ points: [{date, myr_usd}], stale }` — daily USD/MYR series for the chart |
| `PUT /api/conversions/:id` | edit a conversion (validated; 404 if absent) |
| `DELETE /api/conversions/:id` | delete a conversion (204; 404 if absent) |

## Notes

- Rates are mid-market; Moomoo's real quote includes a spread and is slightly worse.
- The assessment ranks today's rate against frankfurter's historical daily series (business days only), fetched live per timeframe — no local rate history is stored.
- Money is stored as float for display this phase; Phase 4 tax math will use a decimal type.
- The FX card shows a chart of MYR→USD over the selected timeframe (frankfurter daily closes) with the live rate marked; conversions can be edited or deleted from the history table.

## Tests

Repo tests need a `fxtracker_test` database:

```bash
docker compose exec -T db psql -U fx -d fxtracker -c "CREATE DATABASE fxtracker_test;"
cd server && go test ./... -race
```

## Configuration

Copy `server/.env.example` and adjust `DATABASE_URL` as needed. For a cloud database (Supabase/Neon), use `sslmode=require`.
