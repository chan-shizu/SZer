import { backendFetchJson } from "./server";

export type Comment = {
  id: number;
  program_id: number;
  user_id: string | null;
  user_name: string | null;
  content: string;
  created_at: string;
  updated_at: string;
};

export type GetCommentsResponse = {
  comments: Comment[];
};

export type PostCommentRequest = {
  content: string;
};

export type PostCommentResponse = {
  comment: Comment;
};

// Goバックエンドのsql.NullString型 { String, Valid } → string|null へ変換
function parseNullString(val: any): string | null {
  if (!val || typeof val !== "object") return null;
  return val.Valid ? val.String : null;
}

export async function getComments(programId: number | string): Promise<GetCommentsResponse> {
  const encodedId = encodeURIComponent(String(programId));
  const res = await backendFetchJson<any>(`/programs/${encodedId}/comments`, { method: "GET", cache: "no-store" });
  // 型変換
  const comments: Comment[] = (res.comments || []).map((c: any) => ({
    id: c.id,
    program_id: c.program_id,
    user_id: parseNullString(c.user_id),
    user_name: parseNullString(c.user_name),
    content: c.content,
    created_at: typeof c.created_at === "string" ? c.created_at : new Date(c.created_at).toISOString(),
    updated_at: typeof c.updated_at === "string" ? c.updated_at : new Date(c.updated_at).toISOString(),
  }));
  return { comments };
}

export async function postComment(
  programId: number | string,
  req: PostCommentRequest,
  cookie?: string,
): Promise<PostCommentResponse> {
  const encodedId = encodeURIComponent(String(programId));
  const headers: Record<string, string> = { "Content-Type": "application/json" };
  if (cookie) headers["Cookie"] = cookie;
  return backendFetchJson<PostCommentResponse>(`/programs/${encodedId}/comments`, {
    method: "POST",
    headers,
    body: JSON.stringify(req),
  });
}
