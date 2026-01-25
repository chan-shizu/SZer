import { NextResponse, type NextRequest } from "next/server";

import { backendFetch } from "@/lib/api/server";

export async function GET(request: NextRequest) {
  const cookie = request.headers.get("cookie");

  const res = await backendFetch("/me/points", {
    method: "GET",
    headers: {
      ...(cookie ? { Cookie: cookie } : {}),
    },
  });

  const text = await res.text();
  return new NextResponse(text, {
    status: res.status,
    headers: {
      "Content-Type": res.headers.get("content-type") || "application/json",
    },
  });
}
