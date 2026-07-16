import { useQuery } from "@tanstack/react-query";
import { fetchDcaStatus } from "@/api/conversions";
import { Card, CardHeader, CardTitle, CardContent } from "@/components/ui/card";

type BlendedRateCardProps = {
  from: string;
  to: string;
};

export function BlendedRateCard({ from, to }: BlendedRateCardProps) {
  const { data } = useQuery({
    queryKey: ["dcaStatus", from, to],
    queryFn: () => fetchDcaStatus(from, to),
    refetchInterval: 60_000,
    enabled: from !== to,
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
              {data.blended_rate.toFixed(4)} <span className="text-base font-medium text-muted-foreground">{to} per 1 {from}</span>
            </p>
            <p className="mt-1 text-sm text-muted-foreground tabular-nums">
              {data.total_home.toFixed(2)} {from} spent → {data.total_target.toFixed(2)} {to} acquired
            </p>
          </>
        )}
      </CardContent>
    </Card>
  );
}
