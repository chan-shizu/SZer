import { auth } from "@/lib/auth/auth";
import { headers as nextHeaders } from "next/headers";
import Link from "next/link";

export default async function UserStatus() {
  let userName: string | null = null;
  try {
    // Build Headers from Next request context so better-auth can use them.
    const reqHeaders = await nextHeaders();
    const hdrs = new Headers();
    const cookie = reqHeaders.get("cookie");
    if (cookie) {
      hdrs.set("cookie", cookie);
    }

    const session = await auth.api.getSession({ headers: hdrs });
    userName = session?.user?.name || session?.user?.email || null;
  } catch (e) {
    console.error("UserStatus session error:", e);
    userName = null;
  }

  return (
    <div className="text-sm text-zinc-700 dark:text-zinc-200">
      {userName ? <Link href="/mypage/profile">{userName}さま</Link> : <Link href="/login">未ログイン</Link>}
    </div>
  );
}
