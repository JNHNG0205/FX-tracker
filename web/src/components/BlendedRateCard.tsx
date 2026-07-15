import { useQuery } from "@tanstack/react-query";
import { fetchDcaStatus } from "@/api/conversions";
import { Card, CardHeader, CardTitle, CardContent } from "@/components/ui/card";

export function BlendedRateCard() {
  const { data } = useQuery({
    queryKey: ["dcaStatus"],
    queryFn: fetchDcaStatus,
    refetchInterval: 60_000,
  });

  return (
    <Card className="w-full">
      <CardHeader>
        <CardTitle>Blended Rate</CardTitle>
      </CardHeader>
      <CardContent>
        {!data?.has_data ? (
          <p className="text-sm text-muted-foreground">Log a conversion to see your blended rate.</p>
        ) : (
          <>
            <p className="text-3xl font-bold tracking-tight tabular-nums">
              {data.blended_rate.toFixed(4)} <span className="text-base font-medium text-muted-foreground">USD per 1 MYR</span>
            </p>
            <p className="mt-1 text-sm text-muted-foreground tabular-nums">
              {data.total_myr.toFixed(2)} MYR spent → {data.total_usd.toFixed(2)} USD acquired
            </p>
          </>
        )}
      </CardContent>
    </Card>
  );
}
