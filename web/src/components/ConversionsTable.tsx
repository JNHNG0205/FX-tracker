import { useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { AlertTriangle } from "lucide-react";
import {
  fetchConversions,
  updateConversion,
  deleteConversion,
  type Conversion,
} from "@/api/conversions";
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

function EditDialog({ conversion }: { conversion: Conversion }) {
  const queryClient = useQueryClient();
  const [open, setOpen] = useState(false);
  const [from, setFrom] = useState(conversion.from_currency);
  const [to, setTo] = useState(conversion.to_currency);
  const [amount, setAmount] = useState(String(conversion.from_amount));
  const [rate, setRate] = useState(String(conversion.rate));
  const [note, setNote] = useState(conversion.note);

  const mutation = useMutation({
    mutationFn: () =>
      updateConversion(conversion.id, {
        from_currency: from,
        to_currency: to,
        from_amount: Number(amount),
        rate: Number(rate),
        note,
        date: conversion.date,
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["conversions"] });
      queryClient.invalidateQueries({ queryKey: ["dcaStatus"] });
      setOpen(false);
    },
  });

  const samePair = from === to;
  const valid = Number(amount) > 0 && Number(rate) > 0 && !samePair;

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger render={<Button variant="outline" size="sm" />}>Edit</DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Edit conversion</DialogTitle>
        </DialogHeader>
        <form
          className="space-y-3"
          onSubmit={(e) => {
            e.preventDefault();
            if (valid) mutation.mutate();
          }}
        >
          <div className="flex flex-wrap gap-4">
            <div className="flex flex-col gap-1.5">
              <Label>From</Label>
              <CurrencySelect value={from} onChange={setFrom} label="From currency" />
            </div>
            <div className="flex flex-col gap-1.5">
              <Label>To</Label>
              <CurrencySelect value={to} onChange={setTo} label="To currency" />
            </div>
          </div>
          {samePair && (
            <p className="flex items-center gap-1.5 text-sm text-caution">
              <AlertTriangle className="size-4" aria-hidden="true" />
              From and To currencies must differ.
            </p>
          )}
          <div>
            <Label htmlFor={`amount-${conversion.id}`}>Amount ({from})</Label>
            <Input id={`amount-${conversion.id}`} type="number" step="0.01" min="0" value={amount} onChange={(e) => setAmount(e.target.value)} />
          </div>
          <div>
            <Label htmlFor={`rate-${conversion.id}`}>Rate ({to} per 1 {from})</Label>
            <Input id={`rate-${conversion.id}`} type="number" step="0.0001" min="0" value={rate} onChange={(e) => setRate(e.target.value)} />
          </div>
          <div>
            <Label htmlFor={`note-${conversion.id}`}>Note</Label>
            <Input id={`note-${conversion.id}`} value={note} onChange={(e) => setNote(e.target.value)} />
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
    mutationFn: () => deleteConversion(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["conversions"] });
      queryClient.invalidateQueries({ queryKey: ["dcaStatus"] });
    },
  });
  return (
    <AlertDialog>
      <AlertDialogTrigger render={<Button variant="destructive" size="sm" />}>Delete</AlertDialogTrigger>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>Delete this conversion?</AlertDialogTitle>
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

export function ConversionsTable() {
  const { data, isLoading } = useQuery({ queryKey: ["conversions"], queryFn: fetchConversions });

  return (
    <Card className="w-full">
      <CardHeader>
        <CardTitle>History</CardTitle>
      </CardHeader>
      <CardContent>
        {isLoading ? (
          <p className="text-sm text-muted-foreground">Loading…</p>
        ) : !data || data.length === 0 ? (
          <p className="text-sm text-muted-foreground">No conversions logged yet.</p>
        ) : (
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Date</TableHead>
                <TableHead>Pair</TableHead>
                <TableHead>Amount</TableHead>
                <TableHead>Rate</TableHead>
                <TableHead>Received</TableHead>
                <TableHead>Note</TableHead>
                <TableHead className="text-right">Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {data.map((c) => (
                <TableRow key={c.id}>
                  <TableCell>{c.date.slice(0, 10)}</TableCell>
                  <TableCell>{c.from_currency}→{c.to_currency}</TableCell>
                  <TableCell className="tabular-nums">{c.from_amount.toFixed(2)}</TableCell>
                  <TableCell className="tabular-nums">{c.rate.toFixed(4)}</TableCell>
                  <TableCell className="tabular-nums">{(c.from_amount * c.rate).toFixed(2)}</TableCell>
                  <TableCell
                    className="max-w-[16rem] truncate text-muted-foreground"
                    title={c.note}
                  >
                    {c.note}
                  </TableCell>
                  <TableCell className="flex justify-end gap-2">
                    <EditDialog conversion={c} />
                    <DeleteButton id={c.id} />
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        )}
      </CardContent>
    </Card>
  );
}
