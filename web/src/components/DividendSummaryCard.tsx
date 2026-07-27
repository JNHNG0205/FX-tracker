import { useQuery } from "@tanstack/react-query";
import { fetchDividends } from "@/api/dividends";
import { useActivePair } from "@/lib/pair";
import { Card, CardHeader, CardTitle, CardContent } from "@/components/ui/card";

export function DividendSummaryCard() {
  const { home } = useActivePair();
  const { data } = useQuery({ queryKey: ["dividends", home], queryFn: () => fetchDividends(home) });

  const hasDividends = !!data && data.dividends.length > 0;

  return (
    <Card className="w-full">
      <CardHeader>
        <CardTitle>Dividend Income</CardTitle>
      </CardHeader>
      <CardContent>
        {!hasDividends ? (
          <p className="text-sm text-muted-foreground">Log a dividend to see your net income.</p>
        ) : (
          <SummaryHero
            homeCurrency={data.home_currency}
            homeNet={data.totals.home_net}
            byCurrency={data.totals.by_currency}
          />
        )}
      </CardContent>
    </Card>
  );
}

type SummaryHeroProps = {
  homeCurrency: string;
  homeNet: number;
  byCurrency: { currency: string; gross: string; withholding: string; net: string }[];
};

function SummaryHero({ homeCurrency, homeNet, byCurrency }: SummaryHeroProps) {
  const single = byCurrency.length === 1 ? byCurrency[0] : undefined;

  return (
    <div className="flex flex-col gap-3">
      <div className="flex flex-col gap-1">
        <span className="text-3xl font-bold tabular-nums">
          {homeNet.toFixed(2)} <span className="text-base font-medium text-muted-foreground">{homeCurrency} net income</span>
        </span>
        {single ? (
          <p className="text-sm text-muted-foreground tabular-nums">
            {single.gross} gross − {single.withholding} withholding {single.currency}
          </p>
        ) : (
          <p className="text-sm text-muted-foreground">
            Across {byCurrency.length} currencies — see breakdown below
          </p>
        )}
      </div>
      {byCurrency.length > 1 && (
        <div className="flex flex-col gap-1 border-t pt-2">
          {byCurrency.map((c) => (
            <p key={c.currency} className="text-sm text-muted-foreground tabular-nums">
              {c.currency}: {c.gross} gross − {c.withholding} withholding = {c.net} net
            </p>
          ))}
        </div>
      )}
    </div>
  );
}
