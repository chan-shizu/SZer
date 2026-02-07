import { NextRequest, NextResponse } from "next/server";
import { postComment } from "@/lib/api/comments";

export async function POST(req: NextRequest, context: { params: Promise<{ program_id: string }> }) {
  const { program_id } = await context.params;

  try {
    const body = await req.json();
    // Cookieをバックエンドにforward
    const cookie = req.headers.get("cookie") ?? undefined;
    const data = await postComment(program_id, body, cookie);
    return NextResponse.json(data);
  } catch {
    return NextResponse.json({ error: "コメント投稿に失敗しました" }, { status: 500 });
  }
}
