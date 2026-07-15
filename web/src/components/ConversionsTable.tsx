import { useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  fetchConversions,
  updateConversion,
  deleteConversion,
  type Conversion,
} from "@/api/conversions";
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
  const [myrAmount, setMyrAmount] = useState(String(conversion.myr_amount));
  const [rate, setRate] = useState(String(conversion.rate_myr_usd));
  const [note, setNote] = useState(conversion.note);

  const mutation = useMutation({
    mutationFn: () =>
      updateConversion(conversion.id, {
        myr_amount: Number(myrAmount),
        rate_myr_usd: Number(rate),
        note,
        date: conversion.date,
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["conversions"] });
      queryClient.invalidateQueries({ queryKey: ["dcaStatus"] });
      setOpen(false);
    },
  });

  const valid = Number(myrAmount) > 0 && Number(rate) > 0;

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
          <div>
            <Label htmlFor={`myr-${conversion.id}`}>MYR amount</Label>
            <Input id={`myr-${conversion.id}`} type="number" step="0.01" min="0" value={myrAmount} onChange={(e) => setMyrAmount(e.target.value)} />
          </div>
          <div>
            <Label htmlFor={`rate-${conversion.id}`}>Rate (USD per 1 MYR)</Label>
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
      <AlertDialogTrigger render={<Button variant="outline" size="sm" />}>Delete</AlertDialogTrigger>
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
                <TableHead>MYR</TableHead>
                <TableHead>Rate</TableHead>
                <TableHead>USD</TableHead>
                <TableHead>Note</TableHead>
                <TableHead className="text-right">Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {data.map((c) => (
                <TableRow key={c.id}>
                  <TableCell>{c.date.slice(0, 10)}</TableCell>
                  <TableCell className="tabular-nums">{c.myr_amount.toFixed(2)}</TableCell>
                  <TableCell className="tabular-nums">{c.rate_myr_usd.toFixed(4)}</TableCell>
                  <TableCell className="tabular-nums">{(c.myr_amount * c.rate_myr_usd).toFixed(2)}</TableCell>
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
