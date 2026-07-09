export type Rate = {
  usd_myr: number;
  myr_usd: number;
  fetched_at: string;
  stale: boolean;
};

export type RateContext = {
  current_myr_usd: number;
  min_myr_usd: number;
  max_myr_usd: number;
  assessment: "good" | "middling" | "poor" | "unknown";
  stale: boolean;
  fetched_at: string;
};

async function getJson<T>(url: string): Promise<T> {
  const res = await fetch(url);
  if (!res.ok) {
    throw new Error(`request failed: ${res.status}`);
  }
  return (await res.json()) as T;
}

export function fetchRate(): Promise<Rate> {
  return getJson<Rate>("/api/rate");
}

export function fetchRateContext(): Promise<RateContext> {
  return getJson<RateContext>("/api/rate/context");
}
