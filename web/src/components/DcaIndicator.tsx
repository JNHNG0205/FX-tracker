import { useQuery } from "@tanstack/react-query";
import { TrendingDown, TrendingUp } from "lucide-react";
import { fetchDcaStatus } from "@/api/conversions";
import { Card, CardHeader, CardTitle, CardContent } from "@/components/ui/card";

export function DcaIndicator() {
  const { data } = useQuery({
    queryKey: ["dcaStatus"],
    queryFn: fetchDcaStatus,
    refetchInterval: 60_000,
  });

  if (!data?.has_data) return null;

  const deltaColor = data.beats_avg ? "text-positive" : "text-negative";
  const DeltaIcon = data.beats_avg ? TrendingUp : TrendingDown;

  return (
    <Card className="w-full">
      <CardHeader>
        <CardTitle>DCA Check</CardTitle>
      </CardHeader>
      <CardContent>
        <div className={`flex items-center gap-2 ${deltaColor}`}>
          <DeltaIcon className="size-6" aria-hidden="true" />
          <span className="text-3xl font-bold tabular-nums">
            {Math.abs(data.delta_pct).toFixed(1)}%
          </span>
        </div>
        <p className="text-sm text-muted-foreground">
          {data.beats_avg ? "better" : "worse"} than your average
        </p>
        <p className="mt-2 text-sm text-muted-foreground tabular-nums">
          Live {data.live_rate.toFixed(4)} vs blended {data.blended_rate.toFixed(4)} (USD per 1 MYR)
        </p>
      </CardContent>
    </Card>
  );
}
