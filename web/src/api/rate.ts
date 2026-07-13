export type Rate = {
  usd_myr: number;
  myr_usd: number;
  fetched_at: string;
  stale: boolean;
};

export type Verdict = "good" | "middling" | "poor" | "unknown";

export type TimeframeAssessment = {
  label: string;
  assessment: Verdict;
  percentile: number;
  min: number;
  max: number;
  samples: number;
  start: string;
  end: string;
};

export type RateContext = {
  current_myr_usd: number;
  current_usd_myr: number;
  stale: boolean;
  fetched_at: string;
  history_stale: boolean;
  timeframes: TimeframeAssessment[];
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

export type RatePoint = {
  date: string;
  myr_usd: number;
};

export type RateHistory = {
  points: RatePoint[];
  stale: boolean;
};

export function fetchRateHistory(): Promise<RateHistory> {
  return getJson<RateHistory>("/api/rate/history");
}
