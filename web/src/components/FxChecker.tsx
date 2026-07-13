import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { fetchRate, fetchRateContext, fetchRateHistory, type Verdict } from "@/api/rate";
import {
  Card,
  CardHeader,
  CardTitle,
  CardDescription,
  CardContent,
} from "@/components/ui/card";
import { ToggleGroup, ToggleGroupItem } from "@/components/ui/toggle-group";
import { sliceByTimeframe } from "@/lib/timeframes";
import { RateChart } from "@/components/RateChart";

const assessmentLabel: Record<Verdict, string> = {
  good: "Good time to convert",
  middling: "Middling",
  poor: "Poor time to convert",
  unknown: "Not enough data yet",
};

const assessmentColor: Record<Verdict, string> = {
  good: "text-green-600",
  middling: "text-amber-600",
  poor: "text-red-600",
  unknown: "text-gray-500",
};

const timeframeLabels = ["7d", "14d", "30d", "90d", "YTD"] as const;

export function FxChecker() {
  const [selectedTimeframe, setSelectedTimeframe] = useState<string>("30d");

  const rate = useQuery({ queryKey: ["rate"], queryFn: fetchRate, refetchInterval: 60_000 });
  const ctx = useQuery({
    queryKey: ["rateContext"],
    queryFn: fetchRateContext,
    refetchInterval: 60_000,
  });
  const history = useQuery({ queryKey: ["rateHistory"], queryFn: fetchRateHistory, refetchInterval: 60_000 });

  if (rate.isLoading) return <p>Loading rate…</p>;
  if (rate.isError || !rate.data) return <p>Failed to load rate.</p>;

  const r = rate.data;
  const selected = ctx.data?.timeframes.find((t) => t.label === selectedTimeframe);

  return (
    <Card className="max-w-md">
      <CardHeader>
        <CardTitle>MYR → USD</CardTitle>
        <CardDescription>Live mid-market exchange rate</CardDescription>
      </CardHeader>
      <CardContent>
        <p className="text-3xl font-bold">{r.myr_usd.toFixed(4)} USD per 1 MYR</p>
        <p className="text-sm text-gray-500">USD → MYR: {r.usd_myr.toFixed(4)} MYR per 1 USD</p>

        <ToggleGroup
          className="mt-4"
          value={[selectedTimeframe]}
          onValueChange={(groupValue) => {
            if (groupValue.length > 0) {
              setSelectedTimeframe(groupValue[0]);
            }
          }}
        >
          {timeframeLabels.map((label) => (
            <ToggleGroupItem key={label} value={label} variant="outline">
              {label}
            </ToggleGroupItem>
          ))}
        </ToggleGroup>

        {selected && (
          <div className="mt-3">
            <p className={`font-medium ${assessmentColor[selected.assessment]}`}>
              {assessmentLabel[selected.assessment]}
            </p>
            {selected.samples > 0 ? (
              <p className="text-sm text-gray-500">
                Beats {selected.percentile}% of the last {selected.label} ({selected.samples} days)
              </p>
            ) : (
              <p className="text-sm text-gray-500">No historical data for this window yet.</p>
            )}
            {selected.samples > 0 && selected.samples <= 5 && (
              <p className="text-xs text-gray-400">Few data points — read with caution.</p>
            )}
          </div>
        )}

        {history.data && (
          <RateChart points={sliceByTimeframe(history.data.points, selectedTimeframe)} live={r.myr_usd} />
        )}

        {r.stale && (
          <p className="mt-2 text-sm text-amber-600">
            Showing last known rate — a refresh failed.
          </p>
        )}

        {ctx.data?.history_stale && (
          <p className="mt-1 text-xs text-gray-400">Historical data may be delayed.</p>
        )}

        <p className="mt-4 text-xs text-gray-400">
          Mid-market rate. Moomoo's real quote includes a spread, so it is slightly worse than
          this.
        </p>
      </CardContent>
    </Card>
  );
}
