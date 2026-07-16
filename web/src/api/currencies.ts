export type Currency = {
  code: string;
  name: string;
};

async function getJson<T>(url: string): Promise<T> {
  const res = await fetch(url);
  if (!res.ok) {
    throw new Error(`request failed: ${res.status}`);
  }
  return (await res.json()) as T;
}

export function fetchCurrencies(): Promise<Currency[]> {
  return getJson<Currency[]>("/api/currencies");
}
