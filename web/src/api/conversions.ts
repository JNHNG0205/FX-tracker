export type Conversion = {
  id: number;
  date: string;
  myr_amount: number;
  rate_myr_usd: number;
  note: string;
};

export type NewConversion = {
  date?: string;
  myr_amount: number;
  rate_myr_usd: number;
  note?: string;
};

export type DcaStatus = {
  blended_rate: number;
  live_rate: number;
  beats_avg: boolean;
  delta_pct: number;
  total_myr: number;
  total_usd: number;
  has_data: boolean;
};

async function getJson<T>(url: string): Promise<T> {
  const res = await fetch(url);
  if (!res.ok) throw new Error(`request failed: ${res.status}`);
  return (await res.json()) as T;
}

export function fetchConversions(): Promise<Conversion[]> {
  return getJson<Conversion[]>("/api/conversions");
}

export function fetchDcaStatus(): Promise<DcaStatus> {
  return getJson<DcaStatus>("/api/conversions/status");
}

export async function createConversion(body: NewConversion): Promise<Conversion> {
  const res = await fetch("/api/conversions", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
  });
  if (!res.ok) {
    const data = (await res.json().catch(() => ({}))) as { error?: string };
    throw new Error(data.error ?? `request failed: ${res.status}`);
  }
  return (await res.json()) as Conversion;
}

export async function updateConversion(id: number, body: NewConversion): Promise<Conversion> {
  const res = await fetch(`/api/conversions/${id}`, {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
  });
  if (!res.ok) {
    const data = (await res.json().catch(() => ({}))) as { error?: string };
    throw new Error(data.error ?? `request failed: ${res.status}`);
  }
  return (await res.json()) as Conversion;
}

export async function deleteConversion(id: number): Promise<void> {
  const res = await fetch(`/api/conversions/${id}`, { method: "DELETE" });
  if (!res.ok) {
    throw new Error(`request failed: ${res.status}`);
  }
}
