import { useQuery } from "@tanstack/react-query";
import { fetchDividends } from "@/api/dividends";
import { useActivePair } from "@/lib/pair";
import { Card, CardHeader, CardTitle, CardContent } from "@/components/ui/card";

// Sums decimal money strings ("12.50") without going through float
// arithmetic, since the plan flags any money math that could compound
// error. Assumes two decimal places, matching the backend's money format.
function sumMoneyStrings(values: string[]): string {
  const totalCents = values.reduce((sum, value) => {
    const [wholePart, fractionPart = "0"] = value.split(".");
    const cents = BigInt(wholePart) * 100n + BigInt(fractionPart.padEnd(2, "0").slice(0, 2));
    return sum + cents;
  }, 0n);
  const negative = totalCents < 0n;
  const abs = negative ? -totalCents : totalCents;
  const whole = abs / 100n;
  const fraction = (abs % 100n).toString().padStart(2, "0");
  return `${negative ? "-" : ""}${whole}.${fraction}`;
}

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
  const totalGross = sumMoneyStrings(byCurrency.map((c) => c.gross));
  const totalWithholding = sumMoneyStrings(byCurrency.map((c) => c.withholding));

  return (
    <div className="flex flex-col gap-3">
      <div className="flex flex-col gap-1">
        <span className="text-3xl font-bold tabular-nums">
          {homeNet.toFixed(2)} <span className="text-base font-medium text-muted-foreground">{homeCurrency} net income</span>
        </span>
        <p className="text-sm text-muted-foreground tabular-nums">
          {totalGross} gross − {totalWithholding} withholding
        </p>
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
