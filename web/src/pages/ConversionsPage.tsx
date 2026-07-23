import { BlendedRateCard } from "@/components/BlendedRateCard";
import { DcaIndicator } from "@/components/DcaIndicator";
import { ConversionForm } from "@/components/ConversionForm";
import { ConversionsTable } from "@/components/ConversionsTable";
import { useActivePair } from "@/lib/pair";

export function ConversionsPage() {
  const { home, target } = useActivePair();

  return (
    <div>
      <h1 className="text-2xl font-bold tracking-tight">Conversions</h1>
      <p className="mt-1 text-muted-foreground">
        Your blended average rate and conversion history.
      </p>

      <div className="mt-6 grid gap-6 lg:grid-cols-2">
        <BlendedRateCard from={home} to={target} />
        <DcaIndicator from={home} to={target} />
      </div>

      <div className="mt-6 grid gap-6 lg:grid-cols-2">
        <ConversionForm />
        <ConversionsTable />
      </div>
    </div>
  );
}
