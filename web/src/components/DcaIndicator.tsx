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
        <p className={`flex items-center gap-1.5 font-medium ${deltaColor}`}>
          <DeltaIcon className="size-4" aria-hidden="true" />
          Today is{" "}
          <span className="font-semibold tabular-nums">
            {Math.abs(data.delta_pct).toFixed(1)}%
          </span>{" "}
          {data.beats_avg ? "better" : "worse"} than your average
        </p>
        <p className="mt-1 text-sm text-muted-foreground tabular-nums">
          Live {data.live_rate.toFixed(4)} vs blended {data.blended_rate.toFixed(4)} (USD per MYR)
        </p>
      </CardContent>
    </Card>
  );
}
