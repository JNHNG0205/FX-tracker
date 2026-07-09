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
- **Frontend:** React + Vite + TypeScript (strict), TanStack Query, Tailwind, shadcn/ui
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
| `GET /api/rate/context` | `{ current_myr_usd, min_myr_usd, max_myr_usd, assessment, stale, fetched_at }` |

## Tests

```bash
cd server && go test ./... -race
```

## Notes

- Rates are mid-market; Moomoo's real quote includes a spread and is slightly worse.
- The assessment range currently spans only how long the server has run. Real 30/90-day context arrives in Phase 2 via persisted rate samples.

## Configuration

Copy `server/.env.example` and adjust `DATABASE_URL` as needed. For a cloud database (Supabase/Neon), use `sslmode=require`.
