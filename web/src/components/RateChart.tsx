import { Line, LineChart, XAxis, YAxis, CartesianGrid, ReferenceDot } from "recharts";
import { ChartContainer, ChartTooltip, ChartTooltipContent, type ChartConfig } from "@/components/ui/chart";
import type { RatePoint } from "@/api/rate";

const config = {
  myr_usd: { label: "USD per 1 MYR", color: "var(--chart-1)" },
} satisfies ChartConfig;

export function RateChart({ points, live }: { points: RatePoint[]; live: number | null }) {
  if (points.length === 0) {
    return <p className="mt-3 text-sm text-muted-foreground">No historical data for this window yet.</p>;
  }
  const lastDate = points[points.length - 1].date;
  return (
    <ChartContainer config={config} className="mt-3 h-40 w-full">
      <LineChart data={points} margin={{ left: 4, right: 8, top: 8, bottom: 4 }}>
        <CartesianGrid vertical={false} />
        <XAxis dataKey="date" tickFormatter={(d: string) => d.slice(5)} tickMargin={8} minTickGap={24} />
        <YAxis domain={["auto", "auto"]} width={52} tickFormatter={(v: number) => v.toFixed(4)} />
        <ChartTooltip content={<ChartTooltipContent />} />
        <Line dataKey="myr_usd" type="monotone" stroke="var(--color-myr_usd)" dot={false} strokeWidth={2} />
        {live != null && (
          <ReferenceDot
            x={lastDate}
            y={live}
            r={4}
            fill="var(--color-myr_usd)"
            stroke="white"
            label={{ value: "now", position: "top", fontSize: 10 }}
          />
        )}
      </LineChart>
    </ChartContainer>
  );
}
