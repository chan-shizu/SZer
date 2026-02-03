import { getComments } from "@/lib/api/comments";
import { Comment } from "@/lib/api/comments";

export default async function CommentsList({ programId }: { programId: number }) {
  let comments: Comment[] = [];
  try {
    const res = await getComments(programId);
    comments = res.comments;
  } catch {
    comments = [];
  }

  // 新着順（降順）でソート
  const sorted = [...comments].sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime());
  return (
    <ul className="space-y-4 divide-y divide-gray-200">
      {sorted.length === 0 && <li className="text-gray-500 text-center py-8">まだコメントはありません</li>}
      {sorted.map((c) => (
        <li key={c.id} className="flex items-start gap-3 pb-4">
          <div className="flex-1">
            <div className="flex items-center gap-2 mb-1">
              <span className="font-semibold text-gray-800 text-sm">{c.user_name ? c.user_name : "未ログイン"}</span>
              <span className="text-xs text-gray-400">{new Date(c.created_at).toLocaleString()}</span>
            </div>
            <div className="text-gray-900 text-base break-words whitespace-pre-line">{c.content}</div>
          </div>
        </li>
      ))}
    </ul>
  );
}
