import Link from "next/link";
import { redirect } from "next/navigation";
import { ArrowLeft } from "lucide-react";

import { ApiError } from "@/lib/api/error";
import { getWatchingPrograms, getMyPoints } from "@/lib/api/mypage";

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

  const { points } = await getMyPoints();

  return (
    <div>
      <div className="flex items-center gap-2 p-4">
        <Link href="/mypage/profile" className="text-foreground">
          <ArrowLeft className="h-5 w-5" />
        </Link>
        <h1 className="text-lg font-semibold text-foreground">ポイントチャージ</h1>
      </div>
      <PointsAddClient initialPoints={points} />
    </div>
  );
}
