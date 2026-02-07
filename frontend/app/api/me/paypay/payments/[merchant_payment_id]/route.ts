import { NextResponse, type NextRequest } from "next/server";

import { backendFetch } from "@/lib/api/server";

export async function GET(request: NextRequest, { params }: { params: Promise<{ merchant_payment_id: string }> }) {
  const cookie = request.headers.get("cookie");
  const { merchant_payment_id } = await params;

  const res = await backendFetch(`/me/paypay/payments/${encodeURIComponent(merchant_payment_id)}`, {
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
