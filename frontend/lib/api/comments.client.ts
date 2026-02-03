import { Comment, GetCommentsResponse, PostCommentRequest, PostCommentResponse } from "./comments";

export async function getCommentsClient(programId: number): Promise<GetCommentsResponse> {
  const res = await fetch(`/api/programs/${programId}/comments`, { cache: "no-store", credentials: "include" });
  if (!res.ok) throw new Error("コメント取得に失敗しました");
  return res.json();
}

export async function postCommentClient(programId: number, req: PostCommentRequest): Promise<PostCommentResponse> {
  const res = await fetch(`/api/programs/${programId}/comments`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(req),
    credentials: "include",
  });
  if (!res.ok) throw new Error("コメント投稿に失敗しました");
  return res.json();
}
