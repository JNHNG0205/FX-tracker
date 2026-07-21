export type PortfolioHoldingResult = {
  id: number;
  ticker: string;
  currency: string;
  shares: number;
  avg_cost: number;
  price: number;
  price_source: "stooq" | "manual" | "unavailable";
  cost_c: number;
  value_c: number;
  asset_pnl_pct: number;
  blended: number;
  spot: number;
  home_cost: number;
  home_value: number;
  total_return_pct: number;
  fx_pct: number;
  home_available: boolean;
};

export type PortfolioTotals = {
  home_cost: number;
  home_value: number;
  total_return_pct: number;
  counted: number;
};

export type Portfolio = {
  home_currency: string;
  holdings: PortfolioHoldingResult[];
  totals: PortfolioTotals;
};

export async function fetchPortfolio(home?: string): Promise<Portfolio> {
  const url = home ? `/api/portfolio?home=${encodeURIComponent(home)}` : "/api/portfolio";
  const res = await fetch(url);
  if (!res.ok) throw new Error(`request failed: ${res.status}`);
  return (await res.json()) as Portfolio;
}
