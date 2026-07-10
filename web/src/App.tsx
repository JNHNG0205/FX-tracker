import { FxChecker } from "@/components/FxChecker";
import { BlendedRateCard } from "@/components/BlendedRateCard";
import { DcaIndicator } from "@/components/DcaIndicator";
import { ConversionForm } from "@/components/ConversionForm";
import { ConversionsTable } from "@/components/ConversionsTable";

export default function App() {
  return (
    <main className="min-h-screen bg-gray-50 p-8">
      <h1 className="mb-6 text-2xl font-bold">MYR → USD Tool</h1>
      <div className="flex flex-col gap-6">
        <div className="flex flex-wrap gap-6">
          <FxChecker />
          <BlendedRateCard />
          <DcaIndicator />
          <ConversionForm />
        </div>
        <ConversionsTable />
      </div>
    </main>
  );
}
