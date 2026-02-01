export type UpsertWatchHistoryRequest = {
  program_id: number;
  position_seconds: number;
  is_completed: boolean;
};

export type WatchHistory = {
  user_id: string;
  program_id: number;
  position_seconds: number;
  is_completed: boolean;
  last_watched_at: string;
  created_at: string;
  updated_at: string;
};

export type UpsertWatchHistoryResponse = {
  watch_history: WatchHistory;
};

export async function upsertWatchHistory(req: UpsertWatchHistoryRequest): Promise<UpsertWatchHistoryResponse> {
  const res = await fetch("/api/watch-histories", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(req),
  });

  if (!res.ok) {
    const text = await res.text().catch(() => "");
    console.error(`[API通信エラー] upsertWatchHistory:`, {
      req,
      status: res.status,
      response: text,
    });
    throw new Error(text || `Request failed with status ${res.status}`);
  }

  return (await res.json()) as UpsertWatchHistoryResponse;
}

export function upsertWatchHistoryBeacon(req: UpsertWatchHistoryRequest): boolean {
  if (typeof navigator === "undefined" || typeof navigator.sendBeacon !== "function") {
    return false;
  }
  const blob = new Blob([JSON.stringify(req)], { type: "application/json" });
  return navigator.sendBeacon("/api/watch-histories", blob);
}
