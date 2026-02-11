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

function StatusCard({
  icon,
  title,
  description,
  programId,
  variant,
}: {
  icon: React.ReactNode;
  title: string;
  description: string;
  programId: string;
  variant: "success" | "error" | "pending";
}) {
  const iconBg = {
    success: "bg-green-100 text-green-600",
    error: "bg-red-100 text-destructive",
    pending: "bg-orange-100 text-brand",
  }[variant];

  return (
    <div className="flex min-h-screen items-center justify-center bg-muted px-4">
      <div className="w-full max-w-sm rounded-2xl bg-background p-8 shadow-lg">
        <div className="flex flex-col items-center gap-y-4 text-center">
          <div className={`flex h-16 w-16 items-center justify-center rounded-full ${iconBg}`}>{icon}</div>
          <h1 className="text-xl font-bold text-foreground">{title}</h1>
          <p className="text-sm leading-relaxed text-muted-foreground">{description}</p>
          <Link
            className="mt-2 inline-block w-full rounded-lg bg-brand py-3 text-center font-bold text-white shadow transition hover:bg-orange-700"
            href={`/programs/${programId}`}
          >
            番組ページに戻る
          </Link>
        </div>
      </div>
    </div>
  );
}

const CheckIcon = () => (
  <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor" className="h-8 w-8">
    <path
      fillRule="evenodd"
      d="M19.916 4.626a.75.75 0 0 1 .208 1.04l-9 13.5a.75.75 0 0 1-1.154.114l-6-6a.75.75 0 0 1 1.06-1.06l5.353 5.353 8.493-12.74a.75.75 0 0 1 1.04-.207Z"
      clipRule="evenodd"
    />
  </svg>
);

const ErrorIcon = () => (
  <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor" className="h-8 w-8">
    <path
      fillRule="evenodd"
      d="M9.401 3.003c1.155-2 4.043-2 5.197 0l7.355 12.748c1.154 2-.29 4.5-2.599 4.5H4.645c-2.309 0-3.752-2.5-2.598-4.5L9.4 3.003ZM12 8.25a.75.75 0 0 1 .75.75v3.75a.75.75 0 0 1-1.5 0V9a.75.75 0 0 1 .75-.75Zm0 8.25a.75.75 0 1 0 0-1.5.75.75 0 0 0 0 1.5Z"
      clipRule="evenodd"
    />
  </svg>
);

const ClockIcon = () => (
  <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor" className="h-8 w-8">
    <path
      fillRule="evenodd"
      d="M12 2.25c-5.385 0-9.75 4.365-9.75 9.75s4.365 9.75 9.75 9.75 9.75-4.365 9.75-9.75S17.385 2.25 12 2.25ZM12.75 6a.75.75 0 0 0-1.5 0v6c0 .414.336.75.75.75h4.5a.75.75 0 0 0 0-1.5h-3.75V6Z"
      clipRule="evenodd"
    />
  </svg>
);

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
      <StatusCard
        icon={<ErrorIcon />}
        title="決済情報が見つかりません"
        description="決済情報を取得できませんでした。お手数ですが、番組ページからもう一度お試しください。"
        programId={program_id}
        variant="error"
      />
    );
  }

  const result = await backendFetchJson<{ status: string; granted: boolean; program_id: number }>(
    `/me/paypay/payments/${encodeURIComponent(merchantPaymentId)}`,
    { method: "GET", cache: "no-store" },
  );

  if (result.status === "COMPLETED" && result.granted) {
    redirect(`/programs/${program_id}`);
  }

  if (result.status === "COMPLETED") {
    return (
      <StatusCard
        icon={<CheckIcon />}
        title="決済が完了しました"
        description="お支払いが正常に処理されました。番組ページに戻って視聴を開始してください。"
        programId={program_id}
        variant="success"
      />
    );
  }

  return (
    <StatusCard
      icon={<ClockIcon />}
      title="決済処理中..."
      description="PayPayでの決済がまだ完了していません。しばらく経ってからページを再読み込みしてください。"
      programId={program_id}
      variant="pending"
    />
  );
}
