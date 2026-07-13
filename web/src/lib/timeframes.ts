import type { RatePoint } from "@/api/rate";

// Day counts per timeframe; 0 = year-to-date. Mirrors the backend windowing.
export const TIMEFRAME_DAYS: Record<string, number> = {
  "7d": 7,
  "14d": 14,
  "30d": 30,
  "90d": 90,
  YTD: 0,
};

// sliceByTimeframe returns the points on/after the window start for `label`,
// computed relative to `today` (defaults to now). Points are date strings YYYY-MM-DD.
export function sliceByTimeframe(points: RatePoint[], label: string, today = new Date()): RatePoint[] {
  const days = TIMEFRAME_DAYS[label] ?? 30;
  let start: Date;
  if (days === 0) {
    start = new Date(today.getFullYear(), 0, 1);
  } else {
    start = new Date(today);
    start.setDate(start.getDate() - days);
  }
  const startStr = start.toISOString().slice(0, 10);
  return points.filter((p) => p.date >= startStr);
}
