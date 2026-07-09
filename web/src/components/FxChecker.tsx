import { useQuery } from "@tanstack/react-query";
import { fetchRate, fetchRateContext } from "@/api/rate";
import { Card, CardTitle } from "@/components/ui/card";

const assessmentLabel: Record<string, string> = {
  good: "Good time to convert",
  middling: "Middling",
  poor: "Poor time to convert",
  unknown: "Not enough data yet",
};

export function FxChecker() {
  const rate = useQuery({ queryKey: ["rate"], queryFn: fetchRate, refetchInterval: 60_000 });
  const ctx = useQuery({ queryKey: ["rateContext"], queryFn: fetchRateContext, refetchInterval: 60_000 });

  if (rate.isLoading) return <p>Loading rate…</p>;
  if (rate.isError || !rate.data) return <p>Failed to load rate.</p>;

  const r = rate.data;
  return (
    <Card className="max-w-md">
      <CardTitle>MYR → USD</CardTitle>
      <p className="mt-2 text-3xl font-bold">{r.myr_usd.toFixed(4)} USD per 1 MYR</p>
      <p className="text-sm text-gray-500">USD → MYR: {r.usd_myr.toFixed(4)} MYR per 1 USD</p>

      {ctx.data && (
        <p className="mt-3 font-medium">{assessmentLabel[ctx.data.assessment] ?? ctx.data.assessment}</p>
      )}

      {r.stale && (
        <p className="mt-2 text-sm text-amber-600">Showing last known rate — a refresh failed.</p>
      )}

      <p className="mt-4 text-xs text-gray-400">
        Mid-market rate. Moomoo's real quote includes a spread, so it is slightly worse than this.
      </p>
    </Card>
  );
}
