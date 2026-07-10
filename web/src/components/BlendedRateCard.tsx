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
    <Card className="max-w-md">
      <CardHeader>
        <CardTitle>Blended Rate</CardTitle>
      </CardHeader>
      <CardContent>
        {!data?.has_data ? (
          <p className="text-sm text-gray-500">Log a conversion to see your blended rate.</p>
        ) : (
          <>
            <p className="text-3xl font-bold">{data.blended_rate.toFixed(4)} USD per 1 MYR</p>
            <p className="mt-1 text-sm text-gray-500">
              {data.total_myr.toFixed(2)} MYR spent → {data.total_usd.toFixed(2)} USD acquired
            </p>
          </>
        )}
      </CardContent>
    </Card>
  );
}
