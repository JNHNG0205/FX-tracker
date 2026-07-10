import { useQuery } from "@tanstack/react-query";
import { fetchConversions } from "@/api/conversions";
import { Card, CardHeader, CardTitle, CardContent } from "@/components/ui/card";
import {
  Table,
  TableHeader,
  TableBody,
  TableRow,
  TableHead,
  TableCell,
} from "@/components/ui/table";

export function ConversionsTable() {
  const { data, isLoading } = useQuery({
    queryKey: ["conversions"],
    queryFn: fetchConversions,
  });

  return (
    <Card>
      <CardHeader>
        <CardTitle>Conversion History</CardTitle>
      </CardHeader>
      <CardContent>
        {isLoading ? (
          <p className="text-sm text-gray-500">Loading…</p>
        ) : !data || data.length === 0 ? (
          <p className="text-sm text-gray-500">No conversions logged yet.</p>
        ) : (
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Date</TableHead>
                <TableHead>MYR</TableHead>
                <TableHead>Rate</TableHead>
                <TableHead>USD</TableHead>
                <TableHead>Note</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {data.map((c) => (
                <TableRow key={c.id}>
                  <TableCell>{c.date.slice(0, 10)}</TableCell>
                  <TableCell>{c.myr_amount.toFixed(2)}</TableCell>
                  <TableCell>{c.rate_myr_usd.toFixed(4)}</TableCell>
                  <TableCell>{(c.myr_amount * c.rate_myr_usd).toFixed(2)}</TableCell>
                  <TableCell>{c.note}</TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        )}
      </CardContent>
    </Card>
  );
}
