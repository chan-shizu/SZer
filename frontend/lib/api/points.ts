export type AddPointsResponse = {
  points: number;
};

export async function addPoints(amount: 100 | 500 | 1000): Promise<AddPointsResponse> {
  const res = await fetch("/api/me/points/add", {
    method: "POST",
    credentials: "include",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ amount }),
  });

  if (!res.ok) {
    const text = await res.text().catch(() => "");
    console.error(`[API通信エラー] addPoints:`, { amount, status: res.status, response: text });
    throw new Error(text || `Request failed with status ${res.status}`);
  }

  return (await res.json()) as AddPointsResponse;
}
