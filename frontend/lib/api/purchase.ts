export type PurchaseProgramResponse = {
  points: number;
};

export async function purchaseProgram(programId: number | string): Promise<PurchaseProgramResponse> {
  const encodedId = encodeURIComponent(String(programId));
  const res = await fetch(`/api/programs/${encodedId}/purchase`, {
    method: "POST",
    credentials: "include",
  });

  if (!res.ok) {
    const text = await res.text().catch(() => "");
    throw new Error(text || `Request failed with status ${res.status}`);
  }

  return (await res.json()) as PurchaseProgramResponse;
}
