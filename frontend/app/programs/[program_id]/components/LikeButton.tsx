"use client";

import { useState } from "react";
import { authClient } from "@/lib/auth/auth-client";
import { AuthModal } from "@/components/AuthModal";
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

  const [modalOpen, setModalOpen] = useState(false);

  async function onClick() {
    if (isSaving) return;

    // ログインチェック
    const { data: session } = await authClient.getSession();
    if (!session?.user) {
      setModalOpen(true);
      return;
    }

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
    <div>
      <button
        type="button"
        onClick={onClick}
        disabled={isSaving}
        aria-label={liked ? "いいね解除" : "いいね"}
        className="disabled:opacity-50 flex items-center gap-x-3"
      >
        <ThumbsUp className={liked ? "h-5 w-5 fill-current text-foreground" : "h-5 w-5 text-foreground"} strokeWidth={2} />
        <div className="text-sm text-muted-foreground">いいね数: {likeCount}</div>
      </button>

      <AuthModal open={modalOpen} onClose={() => setModalOpen(false)} />
    </div>
  );
}
