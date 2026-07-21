import { useQuery } from "@tanstack/react-query";
import { TrendingDown, TrendingUp } from "lucide-react";
import { fetchPortfolio } from "@/api/portfolio";
import { Card, CardHeader, CardTitle, CardContent } from "@/components/ui/card";

export function PortfolioCard() {
  const { data } = useQuery({ queryKey: ["portfolio"], queryFn: () => fetchPortfolio() });

  const hasHoldings = !!data && data.holdings.length > 0;

  return (
    <Card className="w-full">
      <CardHeader>
        <CardTitle>True Return</CardTitle>
      </CardHeader>
      <CardContent>
        {!hasHoldings ? (
          <p className="text-sm text-muted-foreground">Add a holding to see your true return.</p>
        ) : (
          <PortfolioHero
            homeCurrency={data.home_currency}
            homeValue={data.totals.home_value}
            totalReturnPct={data.totals.total_return_pct}
            counted={data.totals.counted}
          />
        )}
      </CardContent>
    </Card>
  );
}

type PortfolioHeroProps = {
  homeCurrency: string;
  homeValue: number;
  totalReturnPct: number;
  counted: number;
};

function PortfolioHero({ homeCurrency, homeValue, totalReturnPct, counted }: PortfolioHeroProps) {
  const isPositive = totalReturnPct >= 0;
  const returnColor = isPositive ? "text-positive" : "text-negative";
  const TrendIcon = isPositive ? TrendingUp : TrendingDown;

  return (
    <div className="flex flex-col gap-1">
      <div className={`flex items-center gap-2 ${returnColor}`}>
        <TrendIcon className="size-6" aria-hidden="true" />
        <span className="text-3xl font-bold tabular-nums">
          {isPositive ? "+" : ""}
          {totalReturnPct.toFixed(2)}%
        </span>
      </div>
      <p className="text-sm text-muted-foreground tabular-nums">
        {homeValue.toFixed(2)} {homeCurrency} total value
      </p>
      <p className="text-sm text-muted-foreground">
        Counts {counted} holding{counted === 1 ? "" : "s"} with a known {homeCurrency} return.
      </p>
    </div>
  );
}
