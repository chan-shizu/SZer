import { NextResponse, type NextRequest } from "next/server";

import { backendFetch } from "@/lib/api/server";

export async function POST(request: NextRequest) {
  const body = await request.text();

  const res = await backendFetch("/watch-histories", {
    method: "POST",
    headers: {
      "Content-Type": request.headers.get("content-type") || "application/json",
    },
    body,
  });

  const text = await res.text();
  return new NextResponse(text, {
    status: res.status,
    headers: {
      "Content-Type": res.headers.get("content-type") || "application/json",
    },
  });
}
