import { FxChecker } from "@/components/FxChecker";
import { BlendedRateCard } from "@/components/BlendedRateCard";
import { DcaIndicator } from "@/components/DcaIndicator";
import { ConversionForm } from "@/components/ConversionForm";
import { ConversionsTable } from "@/components/ConversionsTable";

export default function App() {
  return (
    <main className="min-h-screen bg-muted/30">
      <div className="mx-auto max-w-5xl px-4 py-8">
        <h1 className="text-2xl font-bold tracking-tight">MYR → USD Tool</h1>
        <p className="mt-1 text-muted-foreground">
          The MYR + FX + tax layer your broker skips.
        </p>

        <div className="mt-6 grid gap-6 lg:grid-cols-3">
          <div className="lg:col-span-2">
            <FxChecker />
          </div>
          <div className="flex flex-col gap-6">
            <BlendedRateCard />
            <DcaIndicator />
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
