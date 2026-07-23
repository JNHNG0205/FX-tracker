import { HoldingForm } from "@/components/HoldingForm";
import { HoldingsTable } from "@/components/HoldingsTable";
import { PortfolioCard } from "@/components/PortfolioCard";

export function HoldingsPage() {
  return (
    <div className="flex flex-col gap-6">
      <h1 className="text-2xl font-bold tracking-tight">Holdings</h1>
      <PortfolioCard />
      <div className="grid gap-6 lg:grid-cols-[minmax(0,20rem)_1fr]">
        <HoldingForm />
        <HoldingsTable />
      </div>
    </div>
  );
}
