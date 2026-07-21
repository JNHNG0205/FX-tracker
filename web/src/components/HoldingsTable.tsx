import { useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { TrendingDown, TrendingUp } from "lucide-react";
import { updateHolding, deleteHolding } from "@/api/holdings";
import { fetchPortfolio, type PortfolioHoldingResult } from "@/api/portfolio";
import { useActivePair } from "@/lib/pair";
import { CurrencySelect } from "@/components/CurrencySelect";
import { Card, CardHeader, CardTitle, CardContent } from "@/components/ui/card";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Dialog, DialogContent, DialogFooter, DialogHeader, DialogTitle, DialogClose, DialogTrigger,
} from "@/components/ui/dialog";
import {
  AlertDialog, AlertDialogAction, AlertDialogCancel, AlertDialogContent,
  AlertDialogDescription, AlertDialogFooter, AlertDialogHeader, AlertDialogTitle, AlertDialogTrigger,
} from "@/components/ui/alert-dialog";

type EditDialogProps = {
  holding: PortfolioHoldingResult;
  trigger: React.ReactElement;
  focusManualPrice?: boolean;
};

function EditDialog({ holding, trigger, focusManualPrice }: EditDialogProps) {
  const queryClient = useQueryClient();
  const [open, setOpen] = useState(false);
  const [ticker, setTicker] = useState(holding.ticker);
  const [shares, setShares] = useState(String(holding.shares));
  const [avgCost, setAvgCost] = useState(String(holding.avg_cost));
  const [currency, setCurrency] = useState(holding.currency);
  const [manualPrice, setManualPrice] = useState(
    holding.price_source === "manual" ? String(holding.price) : "",
  );

  const mutation = useMutation({
    mutationFn: () =>
      updateHolding(holding.id, {
        ticker: ticker.trim(),
        shares: Number(shares),
        avg_cost: Number(avgCost),
        currency,
        manual_price: manualPrice ? Number(manualPrice) : undefined,
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["holdings"] });
      queryClient.invalidateQueries({ queryKey: ["portfolio"] });
      setOpen(false);
    },
  });

  const valid = ticker.trim() !== "" && Number(shares) > 0 && Number(avgCost) > 0;

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger render={trigger} />
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Edit holding</DialogTitle>
        </DialogHeader>
        <form
          className="space-y-3"
          onSubmit={(e) => {
            e.preventDefault();
            if (valid) mutation.mutate();
          }}
        >
          <div>
            <Label htmlFor={`ticker-${holding.id}`}>Ticker</Label>
            <Input id={`ticker-${holding.id}`} value={ticker} onChange={(e) => setTicker(e.target.value)} />
          </div>
          <div className="flex flex-wrap gap-4">
            <div>
              <Label htmlFor={`shares-${holding.id}`}>Shares</Label>
              <Input
                id={`shares-${holding.id}`}
                type="number"
                step="0.0001"
                min="0"
                value={shares}
                onChange={(e) => setShares(e.target.value)}
              />
            </div>
            <div>
              <Label htmlFor={`avgCost-${holding.id}`}>Avg cost ({currency})</Label>
              <Input
                id={`avgCost-${holding.id}`}
                type="number"
                step="0.01"
                min="0"
                value={avgCost}
                onChange={(e) => setAvgCost(e.target.value)}
              />
            </div>
            <div>
              <Label>Currency</Label>
              <CurrencySelect value={currency} onChange={setCurrency} label="Asset currency" />
            </div>
          </div>
          <div>
            <Label htmlFor={`manualPrice-${holding.id}`}>Manual price (optional)</Label>
            <Input
              id={`manualPrice-${holding.id}`}
              type="number"
              step="0.01"
              min="0"
              value={manualPrice}
              onChange={(e) => setManualPrice(e.target.value)}
              autoFocus={focusManualPrice}
            />
          </div>
          {mutation.isError && <p className="text-sm text-negative">{mutation.error.message}</p>}
          <DialogFooter>
            <DialogClose render={<Button type="button" variant="outline" />}>Cancel</DialogClose>
            <Button type="submit" disabled={!valid || mutation.isPending}>
              {mutation.isPending ? "Saving…" : "Save"}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}

function DeleteButton({ id }: { id: number }) {
  const queryClient = useQueryClient();
  const mutation = useMutation({
    mutationFn: () => deleteHolding(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["holdings"] });
      queryClient.invalidateQueries({ queryKey: ["portfolio"] });
    },
  });
  return (
    <AlertDialog>
      <AlertDialogTrigger render={<Button variant="destructive" size="sm" />}>Delete</AlertDialogTrigger>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>Delete this holding?</AlertDialogTitle>
          <AlertDialogDescription>This can&apos;t be undone.</AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel>Cancel</AlertDialogCancel>
          <AlertDialogAction onClick={() => mutation.mutate()}>Delete</AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  );
}

function HomeReturnCell({ holding, home }: { holding: PortfolioHoldingResult; home: string }) {
  if (!holding.home_available) {
    return (
      <span className="text-sm text-muted-foreground">
        log a conversion in {holding.currency} to see your {home} return
      </span>
    );
  }

  const isPositive = holding.total_return_pct >= 0;
  const returnColor = isPositive ? "text-positive" : "text-negative";
  const TrendIcon = isPositive ? TrendingUp : TrendingDown;

  return (
    <div className="flex flex-col gap-0.5">
      <div className={`flex items-center gap-1 tabular-nums ${returnColor}`}>
        <TrendIcon className="size-4" aria-hidden="true" />
        {isPositive ? "+" : ""}
        {holding.total_return_pct.toFixed(2)}%
      </div>
      <p className="text-xs text-muted-foreground tabular-nums">
        asset {holding.asset_pnl_pct.toFixed(2)}% + FX {holding.fx_pct.toFixed(2)}%
      </p>
    </div>
  );
}

function PriceCell({ holding }: { holding: PortfolioHoldingResult }) {
  if (holding.price_source === "unavailable") {
    return (
      <EditDialog
        holding={holding}
        focusManualPrice
        trigger={
          <Button variant="ghost" size="sm" className="h-auto p-0 text-muted-foreground underline">
            set price
          </Button>
        }
      />
    );
  }
  return <span className="tabular-nums">{holding.price.toFixed(2)}</span>;
}

export function HoldingsTable() {
  const { home } = useActivePair();
  const { data, isLoading } = useQuery({ queryKey: ["portfolio", home], queryFn: () => fetchPortfolio(home) });

  return (
    <Card className="w-full">
      <CardHeader>
        <CardTitle>Holdings</CardTitle>
      </CardHeader>
      <CardContent>
        {isLoading ? (
          <p className="text-sm text-muted-foreground">Loading…</p>
        ) : !data || data.holdings.length === 0 ? (
          <p className="text-sm text-muted-foreground">No holdings yet.</p>
        ) : (
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Ticker</TableHead>
                <TableHead>Currency</TableHead>
                <TableHead>Shares</TableHead>
                <TableHead>Price</TableHead>
                <TableHead>Value</TableHead>
                <TableHead>Asset%</TableHead>
                <TableHead>Home return</TableHead>
                <TableHead className="text-right">Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {data.holdings.map((h) => {
                const assetPositive = h.asset_pnl_pct >= 0;
                return (
                  <TableRow key={h.id}>
                    <TableCell>{h.ticker}</TableCell>
                    <TableCell>{h.currency}</TableCell>
                    <TableCell className="tabular-nums">{h.shares}</TableCell>
                    <TableCell>
                      <PriceCell holding={h} />
                    </TableCell>
                    <TableCell className="tabular-nums">
                      {h.value_c.toFixed(2)} {h.currency}
                    </TableCell>
                    <TableCell className={`tabular-nums ${assetPositive ? "text-positive" : "text-negative"}`}>
                      {assetPositive ? "+" : ""}
                      {h.asset_pnl_pct.toFixed(2)}%
                    </TableCell>
                    <TableCell>
                      <HomeReturnCell holding={h} home={home} />
                    </TableCell>
                    <TableCell className="flex justify-end gap-2">
                      <EditDialog holding={h} trigger={<Button variant="outline" size="sm" />} />
                      <DeleteButton id={h.id} />
                    </TableCell>
                  </TableRow>
                );
              })}
            </TableBody>
          </Table>
        )}
      </CardContent>
    </Card>
  );
}
