export type Dividend = {
  id: number;
  ticker: string;
  currency: string;
  amount: string;
  date: string;
  note: string;
};

export type NewDividend = {
  ticker: string;
  currency: string;
  amount: string;
  date?: string;
  note?: string;
};

export type DividendEntry = {
  id: number;
  ticker: string;
  currency: string;
  date: string;
  note: string;
  amount: string;
  withholding: string;
  net: string;
  home_net: number;
  home_available: boolean;
};

export type CurrencyTotal = {
  currency: string;
  gross: string;
  withholding: string;
  net: string;
};

export type DividendSummary = {
  home_currency: string;
  dividends: DividendEntry[];
  totals: {
    by_currency: CurrencyTotal[];
    home_net: number;
  };
};

export async function fetchDividends(home?: string): Promise<DividendSummary> {
  const url = home ? `/api/dividends?home=${encodeURIComponent(home)}` : "/api/dividends";
  const res = await fetch(url);
  if (!res.ok) throw new Error(`request failed: ${res.status}`);
  return (await res.json()) as DividendSummary;
}

export async function createDividend(body: NewDividend): Promise<Dividend> {
  const res = await fetch("/api/dividends", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
  });
  if (!res.ok) {
    const data = (await res.json().catch(() => ({}))) as { error?: string };
    throw new Error(data.error ?? `request failed: ${res.status}`);
  }
  return (await res.json()) as Dividend;
}

export async function updateDividend(id: number, body: NewDividend): Promise<Dividend> {
  const res = await fetch(`/api/dividends/${id}`, {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
  });
  if (!res.ok) {
    const data = (await res.json().catch(() => ({}))) as { error?: string };
    throw new Error(data.error ?? `request failed: ${res.status}`);
  }
  return (await res.json()) as Dividend;
}

export async function deleteDividend(id: number): Promise<void> {
  const res = await fetch(`/api/dividends/${id}`, { method: "DELETE" });
  if (!res.ok) {
    const data = (await res.json().catch(() => ({}))) as { error?: string };
    throw new Error(data.error ?? `request failed: ${res.status}`);
  }
}
