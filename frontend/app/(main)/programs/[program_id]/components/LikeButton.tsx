"use client";

import { useState } from "react";
import { ThumbsUp } from "lucide-react";

import { likeProgram, unlikeProgram } from "@/lib/api/likes";

type Props = {
  programId: number;
  initialLiked: boolean;
  initialLikeCount: number;
};

export function LikeButton({ programId, initialLiked, initialLikeCount }: Props) {
  const [liked, setLiked] = useState(initialLiked);
  const [likeCount, setLikeCount] = useState(initialLikeCount);
  const [isSaving, setIsSaving] = useState(false);

  async function onClick() {
    if (isSaving) return;
    setIsSaving(true);

    try {
      const res = liked ? await unlikeProgram(programId) : await likeProgram(programId);
      setLiked(res.liked);
      setLikeCount(res.like_count);
    } catch (e) {
      console.error("failed to toggle like", e);
    } finally {
      setIsSaving(false);
    }
  }

  return (
    <div className="flex items-center gap-x-3">
      <button
        type="button"
        onClick={onClick}
        disabled={isSaving}
        aria-label={liked ? "いいね解除" : "いいね"}
        className="disabled:opacity-50"
      >
        <ThumbsUp className={liked ? "h-5 w-5 fill-current text-gray-900" : "h-5 w-5 text-gray-900"} strokeWidth={2} />
      </button>
      <div className="text-sm text-gray-600">いいね数: {likeCount}</div>
    </div>
  );
}
