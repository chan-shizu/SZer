import Link from "next/link";
import { redirect } from "next/navigation";
import { headers as nextHeaders } from "next/headers";
import { ArrowLeft, Plus } from "lucide-react";

import { auth } from "@/lib/auth/auth";
import { getMyPoints } from "@/lib/api/mypage";
import { BottomTabBar } from "@/components/BottomTabBar";

export default async function ProfilePage() {
  const reqHeaders = await nextHeaders();
  const hdrs = new Headers();
  const cookie = reqHeaders.get("cookie");
  if (cookie) {
    hdrs.set("cookie", cookie);
  }

  const session = await auth.api.getSession({ headers: hdrs });
  if (!session?.user) {
    redirect("/login");
  }

  const { points } = await getMyPoints();

  const user = session.user;

  return (
    <div>
      <div className="flex items-center gap-2 p-4">
        <Link href="/mypage" className="text-foreground">
          <ArrowLeft className="h-5 w-5" />
        </Link>
        <h1 className="text-lg font-semibold text-foreground">アカウント情報</h1>
      </div>

      <div className="space-y-3 px-4">
        <div className="border-b border-muted pb-3">
          <p className="text-xs text-muted-foreground">ユーザー名</p>
          <p className="mt-1 text-sm text-foreground">{user.name}</p>
        </div>

        <div className="border-b border-muted pb-3">
          <p className="text-xs text-muted-foreground">メールアドレス</p>
          <p className="mt-1 text-sm text-foreground">{user.email}</p>
        </div>

        <div className="pb-3">
          <p className="text-xs text-muted-foreground">ポイント残高</p>
          <div className="mt-1 flex items-center justify-between">
            <p className="text-sm text-foreground">{points.toLocaleString()} pt</p>
            <Link
              href="/mypage/points"
              className="flex items-center gap-1 rounded border border-muted px-2 py-1 text-xs text-muted-foreground hover:text-foreground"
            >
              <Plus className="h-3 w-3" />
              チャージ
            </Link>
          </div>
        </div>
      </div>

      <BottomTabBar />
    </div>
  );
}
