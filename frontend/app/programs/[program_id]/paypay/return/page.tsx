import Link from "next/link";
import { redirect } from "next/navigation";

import { ApiError } from "@/lib/api/error";
import { getWatchingPrograms } from "@/lib/api/mypage";
import { backendFetchJson } from "@/lib/api/server";

export const dynamic = "force-dynamic";

type Props = {
  params: Promise<{ program_id: string }>;
  searchParams: Promise<{ merchantPaymentId?: string | string[] }>;
};

export default async function Page({ params, searchParams }: Props) {
  const { program_id } = await params;

  try {
    await getWatchingPrograms();
  } catch (err) {
    if (err instanceof ApiError && (err.status === 401 || err.status === 403)) {
      redirect("/login");
    }
    throw err;
  }

  const resolvedSearchParams = await searchParams;
  const merchantPaymentIdRaw = resolvedSearchParams.merchantPaymentId;
  const merchantPaymentId = Array.isArray(merchantPaymentIdRaw) ? merchantPaymentIdRaw[0] : merchantPaymentIdRaw;

  if (!merchantPaymentId) {
    return (
      <div className="space-y-3 p-3">
        <h1 className="text-lg font-semibold text-foreground">PayPay決済</h1>
        <div className="text-sm text-muted-foreground">merchantPaymentId が見つかりませんでした</div>
        <Link className="text-sm text-foreground underline" href={`/programs/${program_id}`}>
          番組ページに戻る
        </Link>
      </div>
    );
  }

  const result = await backendFetchJson<{ status: string; granted: boolean; program_id: number }>(
    `/me/paypay/payments/${encodeURIComponent(merchantPaymentId)}`,
    { method: "GET", cache: "no-store" }
  );

  if (result.status === "COMPLETED" && result.granted) {
    redirect(`/programs/${program_id}`);
  }

  return (
    <div className="space-y-3 p-3">
      <h1 className="text-lg font-semibold text-foreground">PayPay決済</h1>
      <div className="text-sm text-foreground">ステータス: {result.status}</div>
      <div className="text-sm text-muted-foreground">
        {result.status === "COMPLETED"
          ? "決済が完了しました。番組ページに戻ってください。"
          : "決済がまだ完了していません。しばらくお待ちください。"}
      </div>
      <Link className="text-sm text-foreground underline" href={`/programs/${program_id}`}>
        番組ページに戻る
      </Link>
    </div>
  );
}
