import { NextResponse, type NextRequest } from "next/server";

import { backendFetch } from "@/lib/api/server";

export async function POST(request: NextRequest) {
  const cookie = request.headers.get("cookie");

	const body = await request.text();

  const res = await backendFetch("/me/points/add", {
    method: "POST",
    headers: {
      ...(cookie ? { Cookie: cookie } : {}),
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
