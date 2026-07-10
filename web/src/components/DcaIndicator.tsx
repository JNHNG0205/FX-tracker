import { useQuery } from "@tanstack/react-query";
import { fetchDcaStatus } from "@/api/conversions";
import { Card, CardHeader, CardTitle, CardContent } from "@/components/ui/card";

export function DcaIndicator() {
  const { data } = useQuery({
    queryKey: ["dcaStatus"],
    queryFn: fetchDcaStatus,
    refetchInterval: 60_000,
  });

  if (!data?.has_data) return null;

  const deltaColor = data.beats_avg ? "text-green-600" : "text-red-600";

  return (
    <Card className="max-w-md">
      <CardHeader>
        <CardTitle>DCA Check</CardTitle>
      </CardHeader>
      <CardContent>
        <p className={`font-medium ${deltaColor}`}>
          Today is {Math.abs(data.delta_pct).toFixed(1)}%{" "}
          {data.beats_avg ? "better" : "worse"} than your average
        </p>
        <p className="mt-1 text-sm text-gray-500">
          Live {data.live_rate.toFixed(4)} vs blended {data.blended_rate.toFixed(4)} (USD per MYR)
        </p>
      </CardContent>
    </Card>
  );
}
