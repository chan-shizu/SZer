import { NextRequest, NextResponse } from "next/server";
import { postRequest } from "@/lib/api/requests";

export async function POST(req: NextRequest) {
  try {
    const body = await req.json();
    const cookie = req.headers.get("cookie") ?? undefined;
    const data = await postRequest(body, cookie);
    return NextResponse.json(data, { status: 201 });
  } catch {
    return NextResponse.json({ error: "リクエストの送信に失敗しました" }, { status: 500 });
  }
}
