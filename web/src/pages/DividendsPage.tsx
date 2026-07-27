import { DividendForm } from "@/components/DividendForm";
import { DividendsTable } from "@/components/DividendsTable";
import { DividendSummaryCard } from "@/components/DividendSummaryCard";

export function DividendsPage() {
  return (
    <div className="flex flex-col gap-6">
      <h1 className="text-2xl font-bold tracking-tight">Dividends</h1>
      <DividendSummaryCard />
      <div className="grid gap-6 lg:grid-cols-[minmax(0,20rem)_1fr]">
        <DividendForm />
        <DividendsTable />
      </div>
    </div>
  );
}
