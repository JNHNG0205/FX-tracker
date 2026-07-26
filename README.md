# Anchor

A personal finance tool for a Malaysia-based investor who buys foreign assets (VOO, US stocks) through Moomoo, funding in a home currency. It covers the layer broker apps ignore: **FX + tax**.

## Status

Under active development. FX checker, conversion log/DCA planner, holdings/true-return tracking, and the dividend + withholding tracker are in place. The app is a sidebar app-shell (Convert / Conversions / Holdings / Dividends).

## Roadmap

- **Phase 0–1 — FX checker.** Live mid-market rate between a home currency and any target currency, with a good/middling/poor assessment based on the recent range. Auto-refreshing, with graceful degradation to the last-known rate.
- **Phase 2 — Conversion log + DCA planner.** Log conversions between any currency pair, compute a per-pair blended average rate, and compare today's rate against it.
- **Phase 3 — True-return tracker.** Manual holdings, live prices, and real return translated to your home currency using your blended conversion rate — not spot.
- **Phase 4 — Dividend + withholding tracker.** Logged dividends with the 30% US withholding tax applied for Malaysian residents.

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
| `GET /api/holdings` | manual holdings, newest first |
| `POST /api/holdings` | add a holding `{ticker, shares, avg_cost, currency, manual_price?}` (shares/avg_cost must be > 0; currency must be supported; manual_price, if given, must be > 0) |
| `PUT /api/holdings/:id` | edit a holding (validated; 404 if absent) |
| `DELETE /api/holdings/:id` | delete a holding (204; 404 if absent) |
| `GET /api/portfolio?home=` | per-holding + aggregate true return in `home` (defaults to your saved home currency): `{home_currency, holdings: [...], totals: {home_cost, home_value, total_return_pct, counted}}` |
| `POST /api/dividends` | log a dividend `{ticker, currency, amount, date?, note?}` (currency must be supported; amount must be > 0; date defaults to today) |
| `GET /api/dividends?home=` | dividend + withholding summary in `home` (defaults to your saved home currency): `{home_currency, dividends: [...], totals: {by_currency, home_net}}` |
| `PUT /api/dividends/:id` | edit a dividend (validated; 404 if absent) |
| `DELETE /api/dividends/:id` | delete a dividend (204; 404 if absent) |

## Notes

- Rates are mid-market; Moomoo's real quote includes a spread and is slightly worse.
- The assessment ranks today's rate against frankfurter's historical daily series (business days only), fetched live per timeframe — no local rate history is stored.
- FX rates and holdings money are stored as float for display; dividend/withholding tax math (Phase 4) uses an exact decimal type end-to-end, with only the home-currency approximation crossing back into float via the display-only spot rate.
- The FX card shows a chart of the active pair over the selected timeframe (frankfurter daily closes) with the live rate marked; conversions can be edited or deleted from the history table.

### Multi-currency

Set a home currency in Settings and track it against any target currency — not just MYR→USD. Roughly 30 fiat currencies are supported via frankfurter.app's ECB cross-rates (no crypto). Live rates and rate context are fetched on demand and cached in memory (live rate ~60s TTL, history ~daily TTL) — no local rate history is stored. The blended average rate and DCA comparison are computed per currency pair, not globally. Existing MYR→USD conversions from before the multi-currency migration were backfilled automatically and need no manual action.

### Holdings & true return

Add a holding with its ticker, shares, average cost, and the currency it's denominated in. Prices come from Finnhub's free `/quote` API (set `FINNHUB_API_KEY` in `.env`; get a free key at finnhub.io): US tickers are the plain symbol (AAPL, VOO), and intl tickers use Finnhub's own symbol format. Without a key configured, or for a ticker Finnhub can't resolve, the holding falls back to the `manual_price` you enter (`price_source` in the API response tells you which was used).

True return is computed in your home currency and split into two components: your cost, translated at the **blended** rate you actually paid to acquire that currency (from your conversion log), against the current value, translated at the **spot** rate — so the total return breaks down into asset performance (price change in the asset's own currency) and FX effect (movement between your blended rate and today's spot). A holding needs at least one logged conversion from your home currency into its currency before a home-currency figure can be shown; until then it still shows the asset-only return.

### Dividends & withholding

Log a dividend with its ticker, currency, gross amount, and date. Every dividend held in USD is taxed at a flat **30% US withholding rate**; all other currencies are taxed at **0%** (no other treaty rates are modeled yet). Gross, withholding, and net are shown in the dividend's own currency, plus a home-currency approximation (`home_net`) converted at today's spot rate for each entry, summed across entries where a spot rate was available. The withholding/net math (`internal/dividend.Withhold`) uses exact decimal arithmetic throughout, not float — only the home-currency approximation uses the (display-only) float spot rate.

## Tests

Repo tests need a `fxtracker_test` database:

```bash
docker compose exec -T db psql -U fx -d fxtracker -c "CREATE DATABASE fxtracker_test;"
cd server && go test ./... -race
```

## Configuration

Copy `server/.env.example` and adjust `DATABASE_URL` as needed. For a cloud database (Supabase/Neon), use `sslmode=require`.
