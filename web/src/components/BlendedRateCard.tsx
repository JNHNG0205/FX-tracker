import { useQuery } from "@tanstack/react-query";
import { fetchDcaStatus } from "@/api/conversions";
import { Card, CardHeader, CardTitle, CardContent } from "@/components/ui/card";

// TODO(task10-11): hardcoded stopgap pair to keep the build green; the
// dashboard will pass the user's selected home/target currencies instead.
const STOPGAP_FROM = "MYR";
const STOPGAP_TO = "USD";

export function BlendedRateCard() {
  const { data } = useQuery({
    queryKey: ["dcaStatus", STOPGAP_FROM, STOPGAP_TO],
    queryFn: () => fetchDcaStatus(STOPGAP_FROM, STOPGAP_TO),
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
              {data.total_home.toFixed(2)} MYR spent → {data.total_target.toFixed(2)} USD acquired
            </p>
          </>
        )}
      </CardContent>
    </Card>
  );
}
