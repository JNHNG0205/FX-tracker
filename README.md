# FX Tracker

A personal finance tool for a Malaysia-based investor who buys foreign assets (VOO, US stocks) through Moomoo, funding in a home currency. It covers the layer broker apps ignore: **FX + tax**.

## Status

Under active development. Multi-currency FX checker and conversion log/DCA planner are in place.

## Roadmap

- **Phase 0–1 — FX checker.** Live mid-market rate between a home currency and any target currency, with a good/middling/poor assessment based on the recent range. Auto-refreshing, with graceful degradation to the last-known rate.
- **Phase 2 — Conversion log + DCA planner.** Log conversions between any currency pair, compute a per-pair blended average rate, and compare today's rate against it.
- **Phase 3 — True-return tracker.** Manual holdings, live prices, and real return translated to your home currency using your blended conversion rate — not spot.
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

All pair endpoints take `from`/`to` query params (ISO currency codes; must be supported and must differ). `rate` is always target-per-1-home; `inverse` is home-per-1-target.

| Endpoint | Returns |
| --- | --- |
| `GET /health` | `{"status":"ok"}` |
| `GET /api/currencies` | supported currencies `[{code, name}]` (~30 ECB fiat currencies) |
| `GET /api/settings` | `{home_currency}` |
| `PUT /api/settings` | update `{home_currency}` (must be a supported currency) |
| `GET /api/rate?from=&to=` | `{from, to, rate, inverse, fetched_at, stale}` — mid-market live rate |
| `GET /api/rate/context?from=&to=` | live rate + `timeframes` array (7d/14d/30d/90d/YTD), each with a percentile-based `assessment` |
| `GET /api/rate/history?from=&to=` | `{from, to, points: [{date, value}], stale}` — daily series for the chart |
| `POST /api/conversions` | log a conversion `{date?, from_currency, to_currency, from_amount, rate, note?}` (amount & rate must be > 0; currencies must differ) |
| `GET /api/conversions` | conversions, newest first |
| `PUT /api/conversions/:id` | edit a conversion (validated; 404 if absent) |
| `DELETE /api/conversions/:id` | delete a conversion (204; 404 if absent) |
| `GET /api/conversions/status?from=&to=` | per-pair blended average rate + DCA compare vs the live rate |

## Notes

- Rates are mid-market; Moomoo's real quote includes a spread and is slightly worse.
- The assessment ranks today's rate against frankfurter's historical daily series (business days only), fetched live per timeframe — no local rate history is stored.
- Money is stored as float for display this phase; Phase 4 tax math will use a decimal type.
- The FX card shows a chart of the active pair over the selected timeframe (frankfurter daily closes) with the live rate marked; conversions can be edited or deleted from the history table.

### Multi-currency

Set a home currency in Settings and track it against any target currency — not just MYR→USD. Roughly 30 fiat currencies are supported via frankfurter.app's ECB cross-rates (no crypto). Live rates and rate context are fetched on demand and cached in memory (live rate ~60s TTL, history ~daily TTL) — no local rate history is stored. The blended average rate and DCA comparison are computed per currency pair, not globally. Existing MYR→USD conversions from before the multi-currency migration were backfilled automatically and need no manual action.

## Tests

Repo tests need a `fxtracker_test` database:

```bash
docker compose exec -T db psql -U fx -d fxtracker -c "CREATE DATABASE fxtracker_test;"
cd server && go test ./... -race
```

## Configuration

Copy `server/.env.example` and adjust `DATABASE_URL` as needed. For a cloud database (Supabase/Neon), use `sslmode=require`.
