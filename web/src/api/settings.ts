export type Settings = {
  home_currency: string;
};

async function getJson<T>(url: string): Promise<T> {
  const res = await fetch(url);
  if (!res.ok) {
    throw new Error(`request failed: ${res.status}`);
  }
  return (await res.json()) as T;
}

export function fetchSettings(): Promise<Settings> {
  return getJson<Settings>("/api/settings");
}

export async function updateSettings(code: string): Promise<Settings> {
  const res = await fetch("/api/settings", {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ home_currency: code }),
  });
  if (!res.ok) {
    const data = (await res.json().catch(() => ({}))) as { error?: string };
    throw new Error(data.error ?? `request failed: ${res.status}`);
  }
  return (await res.json()) as Settings;
}
