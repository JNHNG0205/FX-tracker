export type Holding = {
  id: number;
  ticker: string;
  shares: number;
  avg_cost: number;
  currency: string;
  manual_price?: number;
};

export type NewHolding = {
  ticker: string;
  shares: number;
  avg_cost: number;
  currency: string;
  manual_price?: number;
};

export function fetchHoldings(): Promise<Holding[]> {
  return fetch("/api/holdings").then(async (res) => {
    if (!res.ok) throw new Error(`request failed: ${res.status}`);
    return (await res.json()) as Holding[];
  });
}

export async function createHolding(body: NewHolding): Promise<Holding> {
  const res = await fetch("/api/holdings", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
  });
  if (!res.ok) {
    const data = (await res.json().catch(() => ({}))) as { error?: string };
    throw new Error(data.error ?? `request failed: ${res.status}`);
  }
  return (await res.json()) as Holding;
}

export async function updateHolding(id: number, body: NewHolding): Promise<Holding> {
  const res = await fetch(`/api/holdings/${id}`, {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
  });
  if (!res.ok) {
    const data = (await res.json().catch(() => ({}))) as { error?: string };
    throw new Error(data.error ?? `request failed: ${res.status}`);
  }
  return (await res.json()) as Holding;
}

export async function deleteHolding(id: number): Promise<void> {
  const res = await fetch(`/api/holdings/${id}`, { method: "DELETE" });
  if (!res.ok) {
    const data = (await res.json().catch(() => ({}))) as { error?: string };
    throw new Error(data.error ?? `request failed: ${res.status}`);
  }
}
