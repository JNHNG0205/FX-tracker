import { ArrowRight } from "lucide-react";
import { FxChecker } from "@/components/FxChecker";
import { BlendedRateCard } from "@/components/BlendedRateCard";
import { DcaIndicator } from "@/components/DcaIndicator";
import { ConversionForm } from "@/components/ConversionForm";
import { ConversionsTable } from "@/components/ConversionsTable";
import { CurrencySelect } from "@/components/CurrencySelect";
import { useActivePair } from "@/lib/pair";

export default function App() {
  const { home, target, setHome, setTarget } = useActivePair();

  return (
    <main className="min-h-screen bg-muted/30">
      <div className="mx-auto max-w-5xl px-4 py-8">
        <h1 className="text-2xl font-bold tracking-tight">FX Tracker</h1>
        <p className="mt-1 text-muted-foreground">
          The FX layer your broker skips.
        </p>

        <div className="mt-4 flex items-center gap-3">
          <CurrencySelect value={home} onChange={setHome} label="Home currency" />
          <ArrowRight className="size-4 shrink-0 text-muted-foreground" aria-hidden="true" />
          <CurrencySelect value={target} onChange={setTarget} label="Target currency" />
        </div>

        <div className="mt-6 grid gap-6 lg:grid-cols-3">
          <div className="lg:col-span-2">
            <FxChecker from={home} to={target} />
          </div>
          <div className="flex flex-col gap-6">
            <BlendedRateCard from={home} to={target} />
            <DcaIndicator from={home} to={target} />
          </div>
        </div>

        <div className="mt-6 grid gap-6 lg:grid-cols-2">
          <ConversionForm />
          <ConversionsTable />
        </div>
      </div>
    </main>
  );
}
