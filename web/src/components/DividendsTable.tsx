import { useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { updateDividend, deleteDividend, fetchDividends, type DividendEntry } from "@/api/dividends";
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
  dividend: DividendEntry;
  trigger: React.ReactElement;
};

function EditDialog({ dividend, trigger }: EditDialogProps) {
  const queryClient = useQueryClient();
  const [open, setOpen] = useState(false);
  const [ticker, setTicker] = useState(dividend.ticker);
  const [currency, setCurrency] = useState(dividend.currency);
  const [amount, setAmount] = useState(dividend.amount);
  const [date, setDate] = useState(dividend.date);
  const [note, setNote] = useState(dividend.note);

  const mutation = useMutation({
    mutationFn: () =>
      updateDividend(dividend.id, {
        ticker: ticker.trim(),
        currency,
        amount,
        date: date || undefined,
        note: note.trim() || undefined,
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["dividends"] });
      setOpen(false);
    },
  });

  const valid = ticker.trim() !== "" && Number(amount) > 0;

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger render={trigger} />
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Edit dividend</DialogTitle>
        </DialogHeader>
        <form
          className="space-y-3"
          onSubmit={(e) => {
            e.preventDefault();
            if (valid) mutation.mutate();
          }}
        >
          <div>
            <Label htmlFor={`ticker-${dividend.id}`}>Ticker</Label>
            <Input id={`ticker-${dividend.id}`} value={ticker} onChange={(e) => setTicker(e.target.value)} />
          </div>
          <div className="flex flex-wrap gap-4">
            <div>
              <Label htmlFor={`amount-${dividend.id}`}>Gross amount ({currency})</Label>
              <Input
                id={`amount-${dividend.id}`}
                type="text"
                inputMode="decimal"
                value={amount}
                onChange={(e) => setAmount(e.target.value)}
              />
            </div>
            <div>
              <Label>Currency</Label>
              <CurrencySelect value={currency} onChange={setCurrency} label="Asset currency" />
            </div>
          </div>
          <div>
            <Label htmlFor={`date-${dividend.id}`}>Date</Label>
            <Input
              id={`date-${dividend.id}`}
              type="date"
              value={date}
              onChange={(e) => setDate(e.target.value)}
            />
          </div>
          <div>
            <Label htmlFor={`note-${dividend.id}`}>Note (optional)</Label>
            <Input id={`note-${dividend.id}`} value={note} onChange={(e) => setNote(e.target.value)} />
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
    mutationFn: () => deleteDividend(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["dividends"] });
    },
  });
  return (
    <AlertDialog>
      <AlertDialogTrigger render={<Button variant="destructive" size="sm" />}>Delete</AlertDialogTrigger>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>Delete this dividend?</AlertDialogTitle>
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

function WithholdingCell({ dividend }: { dividend: DividendEntry }) {
  const isZero = dividend.withholding === "0.00" || dividend.withholding === "0";
  if (isZero) {
    return <span className="text-muted-foreground">—</span>;
  }
  return (
    <span className="tabular-nums">
      {dividend.withholding}
      <span className="ml-1 text-xs text-muted-foreground">30%</span>
    </span>
  );
}

function HomeNetCell({ dividend, home }: { dividend: DividendEntry; home: string }) {
  if (!dividend.home_available) {
    return <span className="text-sm text-muted-foreground">no rate</span>;
  }
  return (
    <span className="tabular-nums">
      {dividend.home_net.toFixed(2)} {home}
    </span>
  );
}

export function DividendsTable() {
  const { home } = useActivePair();
  const { data, isLoading } = useQuery({ queryKey: ["dividends", home], queryFn: () => fetchDividends(home) });

  return (
    <Card className="w-full">
      <CardHeader>
        <CardTitle>Dividend History</CardTitle>
      </CardHeader>
      <CardContent>
        {isLoading ? (
          <p className="text-sm text-muted-foreground">Loading…</p>
        ) : !data || data.dividends.length === 0 ? (
          <p className="text-sm text-muted-foreground">No dividends logged yet.</p>
        ) : (
          <div className="overflow-x-auto">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Ticker</TableHead>
                  <TableHead>Currency</TableHead>
                  <TableHead>Date</TableHead>
                  <TableHead>Gross</TableHead>
                  <TableHead>Withholding</TableHead>
                  <TableHead>Net</TableHead>
                  <TableHead>≈ home</TableHead>
                  <TableHead className="text-right">Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {data.dividends.map((d) => (
                  <TableRow key={d.id}>
                    <TableCell>{d.ticker}</TableCell>
                    <TableCell>{d.currency}</TableCell>
                    <TableCell>{d.date}</TableCell>
                    <TableCell className="tabular-nums">
                      {d.amount} {d.currency}
                    </TableCell>
                    <TableCell>
                      <WithholdingCell dividend={d} />
                    </TableCell>
                    <TableCell className="tabular-nums">
                      {d.net} {d.currency}
                    </TableCell>
                    <TableCell>
                      <HomeNetCell dividend={d} home={home} />
                    </TableCell>
                    <TableCell className="flex justify-end gap-2">
                      <EditDialog
                        dividend={d}
                        trigger={
                          <Button variant="outline" size="sm">
                            Edit
                          </Button>
                        }
                      />
                      <DeleteButton id={d.id} />
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </div>
        )}
      </CardContent>
    </Card>
  );
}
