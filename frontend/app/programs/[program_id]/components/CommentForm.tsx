"use client";
import { useState } from "react";
import { useRouter } from "next/navigation";
import { postCommentClient } from "@/lib/api/comments.client";

type Props = {
  programId: number;
};

export default function CommentForm({ programId }: Props) {
  const [content, setContent] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const router = useRouter();

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!content.trim()) return;
    setLoading(true);
    setError(null);
    try {
      await postCommentClient(programId, { content });
      setContent("");
      router.refresh();
    } catch (_) {
      setError("コメントの投稿に失敗しました");
    } finally {
      setLoading(false);
    }
  };

  return (
    <form onSubmit={handleSubmit} className="flex items-center gap-3 shadow p-2 align-middle">
      <input
        type="text"
        className="flex-1 bg-gray-100 rounded px-4 text-base shadow-sm py-1"
        placeholder="コメントを入力..."
        value={content}
        onChange={(e) => setContent(e.target.value)}
        disabled={loading}
        maxLength={200}
      />
      <button
        type="submit"
        className="disabled:opacity-50 transition flex items-center justify-center h-10"
        disabled={loading || !content.trim()}
        aria-label="コメント送信"
        style={{ minHeight: "2.5rem" }}
      >
        <svg
          xmlns="http://www.w3.org/2000/svg"
          fill="none"
          viewBox="0 0 24 24"
          strokeWidth={1.5}
          stroke="currentColor"
          className="w-6 h-6 text-gray-700"
        >
          <path strokeLinecap="round" strokeLinejoin="round" d="M4.5 19.5l15-7.5-15-7.5v6l10 1.5-10 1.5v6z" />
        </svg>
      </button>
      {error && <div className="text-red-500 ml-2 text-sm">{error}</div>}
    </form>
  );
}
