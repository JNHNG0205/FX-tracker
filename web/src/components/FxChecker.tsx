import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { AlertTriangle, Info, Minus, TrendingDown, TrendingUp, type LucideIcon } from "lucide-react";
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

type VerdictPresentation = {
  label: string;
  colorClass: string;
  icon: LucideIcon;
};

const verdictPresentation: Record<Verdict, VerdictPresentation> = {
  good: { label: "Good time to convert", colorClass: "text-positive", icon: TrendingUp },
  middling: { label: "Middling", colorClass: "text-caution", icon: Minus },
  poor: { label: "Poor time to convert", colorClass: "text-negative", icon: TrendingDown },
  unknown: { label: "Not enough data yet", colorClass: "text-muted-foreground", icon: Info },
};

const timeframeLabels = ["7d", "14d", "30d", "90d", "YTD"] as const;

function formatFreshness(fetchedAt: string): string {
  const seconds = Math.max(0, Math.floor((Date.now() - new Date(fetchedAt).getTime()) / 1000));
  if (seconds < 5) return "just now";
  if (seconds < 60) return `${seconds}s ago`;
  const minutes = Math.floor(seconds / 60);
  if (minutes < 60) return `${minutes}m ago`;
  const hours = Math.floor(minutes / 60);
  return `${hours}h ago`;
}

type FxCheckerProps = {
  from: string;
  to: string;
};

function FxCheckerSkeleton({ from, to }: FxCheckerProps) {
  return (
    <Card className="w-full">
      <CardHeader>
        <CardTitle>{from} → {to}</CardTitle>
        <CardDescription>Live mid-market exchange rate</CardDescription>
      </CardHeader>
      <CardContent>
        <div className="h-10 w-48 rounded bg-muted motion-safe:animate-pulse" />
        <div className="mt-2 h-4 w-40 rounded bg-muted motion-safe:animate-pulse" />
        <div className="mt-4 h-8 w-56 rounded bg-muted motion-safe:animate-pulse" />
        <div className="mt-4 h-5 w-44 rounded bg-muted motion-safe:animate-pulse" />
      </CardContent>
    </Card>
  );
}

export function FxChecker({ from, to }: FxCheckerProps) {
  const [selectedTimeframe, setSelectedTimeframe] = useState<string>("30d");

  const samePair = from === to;

  const rate = useQuery({
    queryKey: ["rate", from, to],
    queryFn: () => fetchRate(from, to),
    refetchInterval: 60_000,
    enabled: !samePair,
  });
  const ctx = useQuery({
    queryKey: ["rateContext", from, to],
    queryFn: () => fetchRateContext(from, to),
    refetchInterval: 60_000,
    enabled: !samePair,
  });
  const history = useQuery({
    queryKey: ["rateHistory", from, to],
    queryFn: () => fetchRateHistory(from, to),
    refetchInterval: 60_000,
    enabled: !samePair,
  });

  if (samePair) {
    return (
      <Card className="w-full">
        <CardHeader>
          <CardTitle>{from} → {to}</CardTitle>
          <CardDescription>Live mid-market exchange rate</CardDescription>
        </CardHeader>
        <CardContent>
          <p className="text-sm text-muted-foreground">Pick two different currencies.</p>
        </CardContent>
      </Card>
    );
  }

  if (rate.isLoading) return <FxCheckerSkeleton from={from} to={to} />;

  if (rate.isError || !rate.data) {
    return (
      <Card className="w-full">
        <CardHeader>
          <CardTitle>{from} → {to}</CardTitle>
          <CardDescription>Live mid-market exchange rate</CardDescription>
        </CardHeader>
        <CardContent>
          <p className="text-negative font-semibold">Couldn't load the rate.</p>
          <p className="mt-1 text-sm text-muted-foreground">It'll retry automatically.</p>
        </CardContent>
      </Card>
    );
  }

  const r = rate.data;
  const selected = ctx.data?.timeframes.find((t) => t.label === selectedTimeframe);
  const verdict = selected ? verdictPresentation[selected.assessment] : null;
  const VerdictIcon = verdict?.icon;

  return (
    <Card className="w-full">
      <CardHeader>
        <CardTitle>{from} → {to}</CardTitle>
        <CardDescription>Live mid-market exchange rate</CardDescription>
      </CardHeader>
      <CardContent>
        <p className="text-4xl font-bold tracking-tight tabular-nums">{r.rate.toFixed(4)}</p>
        <p className="text-sm text-muted-foreground">{to} per 1 {from}</p>
        <p className="mt-1 text-sm text-muted-foreground tabular-nums">
          {to} → {from}: {r.inverse.toFixed(4)} {from} per 1 {to}
        </p>
        <p className="mt-1 text-xs text-muted-foreground">Updated {formatFreshness(r.fetched_at)}</p>

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

        {selected && verdict && VerdictIcon && (
          <div className="mt-3">
            <p className={`flex items-center gap-1.5 text-base font-semibold ${verdict.colorClass}`}>
              <VerdictIcon className="size-4" aria-hidden="true" />
              {verdict.label}
            </p>
            {selected.samples > 0 ? (
              <p className="text-sm text-muted-foreground">
                Beats {selected.percentile}% of the last {selected.label} ({selected.samples} days)
              </p>
            ) : (
              <p className="text-sm text-muted-foreground">No historical data for this window yet.</p>
            )}
            {selected.samples > 0 && selected.samples <= 5 && (
              <p className="text-xs text-muted-foreground">Few data points — read with caution.</p>
            )}
          </div>
        )}

        {history.data && (
          <RateChart
            points={sliceByTimeframe(history.data.points, selectedTimeframe)}
            live={r.rate}
            from={from}
            to={to}
          />
        )}

        {r.stale && (
          <p className="mt-2 flex items-center gap-1.5 text-sm font-medium text-caution">
            <AlertTriangle className="size-4" aria-hidden="true" />
            Showing last known rate — a refresh failed.
          </p>
        )}

        {ctx.data?.history_stale && (
          <p className="mt-1 flex items-center gap-1.5 text-xs text-muted-foreground">
            <AlertTriangle className="size-3" aria-hidden="true" />
            Historical data may be delayed.
          </p>
        )}

        <p className="mt-4 text-xs text-muted-foreground">
          Mid-market rate. Moomoo's real quote includes a spread, so it is slightly worse than
          this.
        </p>
      </CardContent>
    </Card>
  );
}
