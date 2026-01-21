import { NextResponse, type NextRequest } from "next/server";

import { backendFetch } from "@/lib/api/server";

type Params = { params: Promise<{ program_id: string }> };

export async function POST(request: NextRequest, { params }: Params) {
  const { program_id } = await params;
  const encodedId = encodeURIComponent(String(program_id));

  const cookie = request.headers.get("cookie");

  const res = await backendFetch(`/programs/${encodedId}/like`, {
    method: "POST",
    headers: {
      ...(cookie ? { Cookie: cookie } : {}),
      "Content-Type": request.headers.get("content-type") || "application/json",
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

export async function DELETE(_request: NextRequest, { params }: Params) {
  const { program_id } = await params;
  const encodedId = encodeURIComponent(String(program_id));

  const cookie = _request.headers.get("cookie");

  const res = await backendFetch(`/programs/${encodedId}/like`, {
    method: "DELETE",
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
