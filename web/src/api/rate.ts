export type Rate = {
  from: string;
  to: string;
  rate: number;
  inverse: number;
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
  from: string;
  to: string;
  rate: number;
  inverse: number;
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

function pairQuery(from: string, to: string): string {
  return `?from=${encodeURIComponent(from)}&to=${encodeURIComponent(to)}`;
}

export function fetchRate(from: string, to: string): Promise<Rate> {
  return getJson<Rate>(`/api/rate${pairQuery(from, to)}`);
}

export function fetchRateContext(from: string, to: string): Promise<RateContext> {
  return getJson<RateContext>(`/api/rate/context${pairQuery(from, to)}`);
}

export type RatePoint = {
  date: string;
  value: number;
};

export type RateHistory = {
  from: string;
  to: string;
  points: RatePoint[];
  stale: boolean;
};

export function fetchRateHistory(from: string, to: string): Promise<RateHistory> {
  return getJson<RateHistory>(`/api/rate/history${pairQuery(from, to)}`);
}
