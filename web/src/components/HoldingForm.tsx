import { useEffect, useRef, useState } from "react";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { Check } from "lucide-react";
import { createHolding } from "@/api/holdings";
import { useActivePair } from "@/lib/pair";
import { CurrencySelect } from "@/components/CurrencySelect";
import { Card, CardHeader, CardTitle, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";

export function HoldingForm() {
  const queryClient = useQueryClient();
  const { target } = useActivePair();
  const [ticker, setTicker] = useState("");
  const [shares, setShares] = useState("");
  const [avgCost, setAvgCost] = useState("");
  const [currency, setCurrency] = useState(target);
  const [manualPrice, setManualPrice] = useState("");
  const [justSaved, setJustSaved] = useState(false);
  const savedTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  useEffect(() => {
    return () => {
      if (savedTimeoutRef.current) clearTimeout(savedTimeoutRef.current);
    };
  }, []);

  const mutation = useMutation({
    mutationFn: createHolding,
    onSuccess: () => {
      setTicker("");
      setShares("");
      setAvgCost("");
      setManualPrice("");
      queryClient.invalidateQueries({ queryKey: ["holdings"] });
      queryClient.invalidateQueries({ queryKey: ["portfolio"] });
      setJustSaved(true);
      if (savedTimeoutRef.current) clearTimeout(savedTimeoutRef.current);
      savedTimeoutRef.current = setTimeout(() => setJustSaved(false), 2500);
    },
  });

  const sharesValue = Number(shares);
  const avgCostValue = Number(avgCost);
  const manualPriceValue = manualPrice ? Number(manualPrice) : undefined;
  const isDisabled =
    ticker.trim() === "" || !(sharesValue > 0) || !(avgCostValue > 0) || mutation.isPending;

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (isDisabled) return;
    mutation.mutate({
      ticker: ticker.trim(),
      shares: sharesValue,
      avg_cost: avgCostValue,
      currency,
      manual_price: manualPriceValue,
    });
  }

  return (
    <Card className="w-full">
      <CardHeader>
        <CardTitle>Add a Holding</CardTitle>
      </CardHeader>
      <CardContent>
        <form className="flex flex-col gap-4" onSubmit={handleSubmit}>
          <div className="flex flex-col gap-1.5">
            <Label htmlFor="ticker">Ticker</Label>
            <Input
              id="ticker"
              type="text"
              value={ticker}
              onChange={(e) => setTicker(e.target.value)}
              placeholder="VOO"
            />
          </div>
          <div className="flex flex-wrap gap-4">
            <div className="flex flex-col gap-1.5">
              <Label htmlFor="shares">Shares</Label>
              <Input
                id="shares"
                type="number"
                min="0"
                step="0.0001"
                value={shares}
                onChange={(e) => setShares(e.target.value)}
                placeholder="10"
              />
            </div>
            <div className="flex flex-col gap-1.5">
              <Label htmlFor="avgCost">Avg cost ({currency})</Label>
              <Input
                id="avgCost"
                type="number"
                min="0"
                step="0.01"
                value={avgCost}
                onChange={(e) => setAvgCost(e.target.value)}
                placeholder="400.00"
              />
            </div>
            <div className="flex flex-col gap-1.5">
              <Label>Currency</Label>
              <CurrencySelect value={currency} onChange={setCurrency} label="Asset currency" />
            </div>
          </div>
          <div className="flex flex-col gap-1.5">
            <Label htmlFor="manualPrice">Manual price (optional)</Label>
            <Input
              id="manualPrice"
              type="number"
              min="0"
              step="0.01"
              value={manualPrice}
              onChange={(e) => setManualPrice(e.target.value)}
              placeholder="Only if live price is unavailable"
            />
          </div>
          <Button type="submit" disabled={isDisabled}>
            {mutation.isPending ? "Saving…" : "Add holding"}
          </Button>
          {mutation.isError && (
            <p className="text-sm text-negative">{mutation.error.message}</p>
          )}
          {justSaved && (
            <p className="flex items-center gap-1.5 text-sm text-positive motion-safe:animate-in motion-safe:fade-in">
              <Check className="size-4" aria-hidden="true" />
              Saved
            </p>
          )}
        </form>
      </CardContent>
    </Card>
  );
}
