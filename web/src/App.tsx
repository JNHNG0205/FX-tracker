import { FxChecker } from "@/components/FxChecker";

export default function App() {
  return (
    <main className="min-h-screen bg-gray-50 p-8">
      <h1 className="mb-6 text-2xl font-bold">MYR → USD Tool</h1>
      <FxChecker />
    </main>
  );
}
