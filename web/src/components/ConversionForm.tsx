import { useEffect, useRef, useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { AlertTriangle, Check } from "lucide-react";
import { createConversion } from "@/api/conversions";
import { fetchRate } from "@/api/rate";
import { Card, CardHeader, CardTitle, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";

const INVERSE_RATE_TOLERANCE = 0.15;

export function ConversionForm() {
  const queryClient = useQueryClient();
  const [myrAmount, setMyrAmount] = useState("");
  const [rate, setRate] = useState("");
  const [note, setNote] = useState("");
  const [justSaved, setJustSaved] = useState(false);
  const savedTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  useEffect(() => {
    return () => {
      if (savedTimeoutRef.current) clearTimeout(savedTimeoutRef.current);
    };
  }, []);

  const liveRate = useQuery({ queryKey: ["rate"], queryFn: fetchRate });

  const mutation = useMutation({
    mutationFn: createConversion,
    onSuccess: () => {
      setMyrAmount("");
      setRate("");
      setNote("");
      queryClient.invalidateQueries({ queryKey: ["conversions"] });
      queryClient.invalidateQueries({ queryKey: ["dcaStatus"] });
      setJustSaved(true);
      if (savedTimeoutRef.current) clearTimeout(savedTimeoutRef.current);
      savedTimeoutRef.current = setTimeout(() => setJustSaved(false), 2500);
    },
  });

  const myrAmountValue = Number(myrAmount);
  const rateValue = Number(rate);
  const isDisabled =
    !(myrAmountValue > 0) || !(rateValue > 0) || mutation.isPending;

  const live = liveRate.data?.myr_usd;
  const rateRatio = live && rateValue > 0 ? rateValue / live : null;
  const isFarFromLive =
    rateRatio !== null && (rateRatio > 1 + INVERSE_RATE_TOLERANCE || rateRatio < 1 - INVERSE_RATE_TOLERANCE);

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (isDisabled) return;
    mutation.mutate({
      myr_amount: myrAmountValue,
      rate_myr_usd: rateValue,
      note: note || undefined,
    });
  }

  return (
    <Card className="w-full">
      <CardHeader>
        <CardTitle>Log a Conversion</CardTitle>
      </CardHeader>
      <CardContent>
        <form className="flex flex-col gap-4" onSubmit={handleSubmit}>
          <div className="flex flex-col gap-1.5">
            <Label htmlFor="myr-amount">MYR amount</Label>
            <Input
              id="myr-amount"
              type="number"
              min="0"
              step="0.01"
              value={myrAmount}
              onChange={(e) => setMyrAmount(e.target.value)}
              placeholder="1000"
            />
          </div>
          <div className="flex flex-col gap-1.5">
            <Label htmlFor="rate">Rate received (USD per 1 MYR)</Label>
            <Input
              id="rate"
              type="number"
              min="0"
              step="0.0001"
              value={rate}
              onChange={(e) => setRate(e.target.value)}
              placeholder="0.2130"
            />
            {isFarFromLive && live !== undefined && (
              <p className="flex items-center gap-1.5 text-sm text-caution">
                <AlertTriangle className="size-4" aria-hidden="true" />
                Far from today's live rate ({live.toFixed(4)}). Did you mean the inverse?
              </p>
            )}
          </div>
          <div className="flex flex-col gap-1.5">
            <Label htmlFor="note">Note (optional)</Label>
            <Input
              id="note"
              type="text"
              value={note}
              onChange={(e) => setNote(e.target.value)}
              placeholder="Moomoo top-up"
            />
          </div>
          <Button type="submit" disabled={isDisabled}>
            {mutation.isPending ? "Saving…" : "Log conversion"}
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
