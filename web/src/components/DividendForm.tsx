import { useEffect, useRef, useState } from "react";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { Check } from "lucide-react";
import { createDividend } from "@/api/dividends";
import { useActivePair } from "@/lib/pair";
import { CurrencySelect } from "@/components/CurrencySelect";
import { Card, CardHeader, CardTitle, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";

export function DividendForm() {
  const queryClient = useQueryClient();
  const { target } = useActivePair();
  const [ticker, setTicker] = useState("");
  const [currency, setCurrency] = useState(target);
  const [amount, setAmount] = useState("");
  const [date, setDate] = useState("");
  const [note, setNote] = useState("");
  const [justSaved, setJustSaved] = useState(false);
  const savedTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  useEffect(() => {
    return () => {
      if (savedTimeoutRef.current) clearTimeout(savedTimeoutRef.current);
    };
  }, []);

  const mutation = useMutation({
    mutationFn: createDividend,
    onSuccess: () => {
      setTicker("");
      setAmount("");
      setDate("");
      setNote("");
      queryClient.invalidateQueries({ queryKey: ["dividends"] });
      setJustSaved(true);
      if (savedTimeoutRef.current) clearTimeout(savedTimeoutRef.current);
      savedTimeoutRef.current = setTimeout(() => setJustSaved(false), 2500);
    },
  });

  const amountValue = Number(amount);
  const isDisabled = ticker.trim() === "" || !(amountValue > 0) || mutation.isPending;

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (isDisabled) return;
    mutation.mutate({
      ticker: ticker.trim(),
      currency,
      amount,
      date: date || undefined,
      note: note.trim() || undefined,
    });
  }

  return (
    <Card className="w-full">
      <CardHeader>
        <CardTitle>Log a Dividend</CardTitle>
      </CardHeader>
      <CardContent>
        <form className="flex flex-col gap-4" onSubmit={handleSubmit}>
          <div className="flex flex-col gap-1.5">
            <Label htmlFor="dividendTicker">Ticker</Label>
            <Input
              id="dividendTicker"
              type="text"
              value={ticker}
              onChange={(e) => setTicker(e.target.value)}
              placeholder="VOO"
            />
          </div>
          <div className="flex flex-wrap gap-4">
            <div className="flex flex-col gap-1.5">
              <Label htmlFor="dividendAmount">Gross amount ({currency})</Label>
              <Input
                id="dividendAmount"
                type="text"
                inputMode="decimal"
                value={amount}
                onChange={(e) => setAmount(e.target.value)}
                placeholder="25.00"
              />
            </div>
            <div className="flex flex-col gap-1.5">
              <Label>Currency</Label>
              <CurrencySelect value={currency} onChange={setCurrency} label="Asset currency" />
            </div>
          </div>
          <div className="flex flex-col gap-1.5">
            <Label htmlFor="dividendDate">Date</Label>
            <Input id="dividendDate" type="date" value={date} onChange={(e) => setDate(e.target.value)} />
          </div>
          <div className="flex flex-col gap-1.5">
            <Label htmlFor="dividendNote">Note (optional)</Label>
            <Input
              id="dividendNote"
              type="text"
              value={note}
              onChange={(e) => setNote(e.target.value)}
              placeholder="Q2 dividend"
            />
          </div>
          <Button type="submit" disabled={isDisabled}>
            {mutation.isPending ? "Saving…" : "Add dividend"}
          </Button>
          {mutation.isError && <p className="text-sm text-negative">{mutation.error.message}</p>}
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
