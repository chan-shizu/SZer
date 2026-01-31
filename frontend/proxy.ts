import { NextResponse, type NextRequest } from "next/server";

import { auth } from "@/lib/auth/auth";

export async function proxy(request: NextRequest) {
  const { pathname } = request.nextUrl;

  // Public paths
  if (
    pathname.startsWith("/api/auth/") ||
    pathname === "/login" ||
    pathname === "/signup" ||
    pathname.startsWith("/_next/") ||
    pathname === "/favicon.ico"
  ) {
    return NextResponse.next();
  }

  // 認証が必要なのは /mypage 配下ページと /api/me 配下APIのみ
  const needsAuth = pathname.startsWith("/mypage") || pathname.startsWith("/api/me");
  if (!needsAuth) {
    return NextResponse.next();
  }

  let session: unknown;
  try {
    session = await auth.api.getSession({ headers: request.headers });
  } catch {
    session = null;
  }

  // better-auth returns null when not authenticated
  if (!session) {
    // For API requests, respond 401; for pages, redirect to /login
    if (pathname.startsWith("/api/")) {
      return Response.json({ error: "unauthorized" }, { status: 401 });
    }
    return NextResponse.redirect(new URL("/login", request.url));
  }

  return NextResponse.next();
}

export const config = {
  matcher: [
    // /mypage配下と/api/me配下のみ認証ガード
    "/mypage/:path*",
    "/api/me/:path*",
  ],
};
