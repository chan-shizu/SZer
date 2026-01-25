import { redirect } from "next/navigation";

import { ApiError } from "@/lib/api/error";
import { getWatchingPrograms } from "@/lib/api/mypage";
import { backendFetchJson } from "@/lib/api/server";

import { PointsAddClient } from "./PointsAddClient";

export const dynamic = "force-dynamic";

export default async function Page() {
  try {
    // auth必須要件のため、軽いAPI呼び出しで未ログインを弾く
    await getWatchingPrograms();
  } catch (err) {
    if (err instanceof ApiError && (err.status === 401 || err.status === 403)) {
      redirect("/login");
    }
    throw err;
  }

  const pointsRes = await backendFetchJson<{ points: number }>("/me/points", { method: "GET", cache: "no-store" });

  return (
    <div>
      <div className="p-3">
        <h1 className="text-lg font-semibold text-foreground">ポイント追加</h1>
      </div>
      <PointsAddClient initialPoints={pointsRes.points} />
    </div>
  );
}
