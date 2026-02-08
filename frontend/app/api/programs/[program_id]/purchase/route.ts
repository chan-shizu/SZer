import { NextResponse, type NextRequest } from "next/server";

import { backendFetch } from "@/lib/api/server";

type Params = { params: Promise<{ program_id: string }> };

export async function POST(request: NextRequest, { params }: Params) {
  const { program_id } = await params;
  const encodedId = encodeURIComponent(String(program_id));

  const cookie = request.headers.get("cookie");

  const res = await backendFetch(`/programs/${encodedId}/purchase`, {
    method: "POST",
    headers: {
      ...(cookie ? { Cookie: cookie } : {}),
      "Content-Type": "application/json",
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
