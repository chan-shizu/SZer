import Link from "next/link";
import { redirect } from "next/navigation";

import { ApiError } from "@/lib/api/error";
import { getWatchingPrograms } from "@/lib/api/mypage";
import { backendFetchJson } from "@/lib/api/server";

export const dynamic = "force-dynamic";

export default async function Page({ searchParams }: { searchParams: { merchantPaymentId?: string | string[] } }) {
  try {
    // auth必須要件のため、軽いAPI呼び出しで未ログインを弾く
    await getWatchingPrograms();
  } catch (err) {
    if (err instanceof ApiError && (err.status === 401 || err.status === 403)) {
      redirect("/login");
    }
    throw err;
  }

  const merchantPaymentIdRaw = searchParams.merchantPaymentId;
  const merchantPaymentId = Array.isArray(merchantPaymentIdRaw) ? merchantPaymentIdRaw[0] : merchantPaymentIdRaw;

  if (!merchantPaymentId) {
    return (
      <div className="space-y-3 p-3">
        <h1 className="text-lg font-semibold text-foreground">PayPay決済</h1>
        <div className="text-sm text-muted-foreground">merchantPaymentId が見つからなかったよ</div>
        <Link className="text-sm text-foreground underline" href="/mypage/points">
          ポイント画面に戻る
        </Link>
      </div>
    );
  }

  const result = await backendFetchJson<{ status: string; credited: boolean; points: number }>(
    `/me/paypay/payments/${encodeURIComponent(merchantPaymentId)}`,
    { method: "GET", cache: "no-store" }
  );

  return (
    <div className="space-y-3 p-3">
      <h1 className="text-lg font-semibold text-foreground">PayPay決済</h1>
      <div className="text-sm text-foreground">ステータス: {result.status}</div>
      <div className="text-sm text-foreground">ポイント: {result.points}</div>
      <div className="text-sm text-muted-foreground">
        {result.credited ? "ポイント付与したよ" : "まだ付与されてないっぽい"}
      </div>
      <Link className="text-sm text-foreground underline" href="/mypage/points">
        ポイント画面に戻る
      </Link>
    </div>
  );
}
