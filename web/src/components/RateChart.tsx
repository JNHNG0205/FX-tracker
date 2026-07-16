import { Line, LineChart, XAxis, YAxis, CartesianGrid, ReferenceDot } from "recharts";
import { ChartContainer, ChartTooltip, ChartTooltipContent, type ChartConfig } from "@/components/ui/chart";
import type { RatePoint } from "@/api/rate";

const config = {
  value: { label: "USD per 1 MYR", color: "var(--chart-1)" },
} satisfies ChartConfig;

export function RateChart({ points, live }: { points: RatePoint[]; live: number | null }) {
  if (points.length === 0) {
    return <p className="mt-3 text-sm text-muted-foreground">No historical data for this window yet.</p>;
  }
  const firstDate = points[0].date;
  const lastDate = points[points.length - 1].date;
  const values = points.map((p) => p.value).concat(live != null ? [live] : []);
  const rawMin = Math.min(...values);
  const rawMax = Math.max(...values);
  const range = rawMax - rawMin;
  const padding = range > 0 ? range * 0.02 : 0.0001;
  const domain: [number, number] = [rawMin - padding, rawMax + padding];
  const summary =
    `MYR to USD over the selected window: ${points.length} points from ${firstDate} to ${lastDate}` +
    (live != null ? `, currently ${live.toFixed(4)}` : "");
  return (
    <div role="img" aria-label={summary} className="mt-3">
      <span className="sr-only">{summary}</span>
      <ChartContainer config={config} className="h-40 w-full" aria-hidden="true">
        <LineChart data={points} margin={{ left: 4, right: 8, top: 8, bottom: 4 }}>
          <CartesianGrid vertical={false} />
          <XAxis dataKey="date" tickFormatter={(d: string) => d.slice(5)} tickMargin={8} minTickGap={24} />
          <YAxis domain={domain} width={52} tickFormatter={(v: number) => v.toFixed(4)} />
          <ChartTooltip content={<ChartTooltipContent />} />
          <Line dataKey="value" type="monotone" stroke="var(--color-primary)" dot={false} strokeWidth={2} />
          {live != null && (
            <ReferenceDot
              x={lastDate}
              y={live}
              r={4}
              fill="var(--color-foreground)"
              stroke="var(--color-background)"
              label={{ value: "now", position: "top", fontSize: 10 }}
            />
          )}
        </LineChart>
      </ChartContainer>
    </div>
  );
}
