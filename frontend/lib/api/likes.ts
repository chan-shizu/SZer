export type LikeResponse = {
  liked: boolean;
  like_count: number;
};

export async function likeProgram(programId: number | string): Promise<LikeResponse> {
  const encodedId = encodeURIComponent(String(programId));
  const res = await fetch(`/api/programs/${encodedId}/like`, { method: "POST", credentials: "include" });

  if (!res.ok) {
    const text = await res.text().catch(() => "");
    console.error(`[API通信エラー] likeProgram:`, { programId, status: res.status, response: text });
    throw new Error(text || `Request failed with status ${res.status}`);
  }

  return (await res.json()) as LikeResponse;
}

export async function unlikeProgram(programId: number | string): Promise<LikeResponse> {
  const encodedId = encodeURIComponent(String(programId));
  const res = await fetch(`/api/programs/${encodedId}/like`, { method: "DELETE", credentials: "include" });

  if (!res.ok) {
    const text = await res.text().catch(() => "");
    console.error(`[API通信エラー] unlikeProgram:`, { programId, status: res.status, response: text });
    throw new Error(text || `Request failed with status ${res.status}`);
  }

  return (await res.json()) as LikeResponse;
}
