import { useState } from "react";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { createConversion } from "@/api/conversions";
import { Card, CardHeader, CardTitle, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";

export function ConversionForm() {
  const queryClient = useQueryClient();
  const [myrAmount, setMyrAmount] = useState("");
  const [rate, setRate] = useState("");
  const [note, setNote] = useState("");

  const mutation = useMutation({
    mutationFn: createConversion,
    onSuccess: () => {
      setMyrAmount("");
      setRate("");
      setNote("");
      queryClient.invalidateQueries({ queryKey: ["conversions"] });
      queryClient.invalidateQueries({ queryKey: ["dcaStatus"] });
    },
  });

  const myrAmountValue = Number(myrAmount);
  const rateValue = Number(rate);
  const isDisabled =
    !(myrAmountValue > 0) || !(rateValue > 0) || mutation.isPending;

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
    <Card className="max-w-md">
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
            <Label htmlFor="rate">Rate received (USD per MYR)</Label>
            <Input
              id="rate"
              type="number"
              min="0"
              step="0.0001"
              value={rate}
              onChange={(e) => setRate(e.target.value)}
              placeholder="0.2130"
            />
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
            <p className="text-sm text-red-600">{mutation.error.message}</p>
          )}
        </form>
      </CardContent>
    </Card>
  );
}
