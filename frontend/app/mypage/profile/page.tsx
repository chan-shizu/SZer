import Link from "next/link";
import { redirect } from "next/navigation";
import { headers as nextHeaders } from "next/headers";
import { ArrowLeft } from "lucide-react";

import { auth } from "@/lib/auth/auth";
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

  const user = session.user;

  return (
    <div>
      <div className="flex items-center gap-2 p-4">
        <Link href="/mypage" className="text-foreground">
          <ArrowLeft className="h-5 w-5" />
        </Link>
        <h1 className="text-lg font-semibold text-foreground">アカウント情報</h1>
      </div>

      <div className="space-y-4 px-4">
        <div className="border-b border-muted pb-4">
          <p className="text-sm text-muted-foreground">ユーザー名</p>
          <p className="mt-1 text-base text-foreground">{user.name}</p>
        </div>

        <div className="pb-4">
          <p className="text-sm text-muted-foreground">メールアドレス</p>
          <p className="mt-1 text-base text-foreground">{user.email}</p>
        </div>
      </div>

      <BottomTabBar />
    </div>
  );
}
