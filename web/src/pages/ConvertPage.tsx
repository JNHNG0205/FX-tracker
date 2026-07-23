import { FxChecker } from "@/components/FxChecker";
import { CurrencySelect } from "@/components/CurrencySelect";
import { useActivePair } from "@/lib/pair";

export function ConvertPage() {
  const { home, target, setTarget } = useActivePair();

  return (
    <div>
      <h1 className="text-2xl font-bold tracking-tight">Convert</h1>
      <p className="mt-1 text-muted-foreground">
        The FX layer your broker skips.
      </p>

      <div className="mt-4">
        <CurrencySelect value={target} onChange={setTarget} label="Target currency" />
      </div>

      <div className="mt-6">
        <FxChecker from={home} to={target} />
      </div>
    </div>
  );
}
