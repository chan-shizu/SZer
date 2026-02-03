"use client";

import { useEffect, useRef } from "react";
import { useRouter } from "next/navigation";

import { upsertWatchHistory, upsertWatchHistoryBeacon } from "@/lib/api/watch_histories";
import { authClient } from "@/lib/auth/auth-client";

type Props = {
  programId: number;
  videoUrl: string;
  startPositionSeconds?: number;
};

export const Video = ({ programId, videoUrl, startPositionSeconds }: Props) => {
  const ref = useRef<HTMLVideoElement | null>(null);
  const sentRef = useRef(false);
  const appliedStartRef = useRef(false);
  const router = useRouter();

  useEffect(() => {
    const video = ref.current;
    if (!video) return;

    appliedStartRef.current = false;

    const start =
      typeof startPositionSeconds === "number" && Number.isFinite(startPositionSeconds) ? startPositionSeconds : 0;
    let seekTarget: number | null = null;

    const computeSeekTarget = (): number | null => {
      if (appliedStartRef.current) return null;
      if (!(start > 0)) return null;

      const duration = video.duration;
      if (!Number.isFinite(duration) || duration <= 0) return null;
      return Math.min(Math.max(start, 0), Math.max(duration - 0.1, 0));
    };

    const tryApplyStart = () => {
      if (appliedStartRef.current) return;

      const target = computeSeekTarget();
      if (target === null) return;

      seekTarget = target;
      try {
        video.currentTime = target;
      } catch {
        // Some browsers throw if the media isn't seekable yet.
      }
    };

    const onSeeked = () => {
      if (appliedStartRef.current) return;
      if (seekTarget === null) return;
      // Mark as applied once we land close enough.
      if (Math.abs(video.currentTime - seekTarget) < 1) {
        appliedStartRef.current = true;
      }
    };

    const send = async (isCompleted: boolean) => {
      if (sentRef.current) return;
      const positionSeconds = Math.floor(video.currentTime || 0);
      if (positionSeconds < 1.0 && !isCompleted) {
        // do not send very beginning position to reduce API calls
        return;
      }

      // postgresのconflictが発生するため、開始位置と同じ位置かつ未完了の場合は送信しない
      if (positionSeconds == startPositionSeconds) return;

      const { data: session } = await authClient.getSession();
      if (!session?.user) {
        // not logged in
        return;
      }

      sentRef.current = true;
      const payload = {
        program_id: programId,
        position_seconds: Math.max(positionSeconds, 0),
        is_completed: isCompleted,
      };

      if (!upsertWatchHistoryBeacon(payload)) {
        void upsertWatchHistory(payload).catch(() => {
          // ignore
        });
      }
    };

    const onEnded = () => send(true);

    // Page leave signals (browser back / close tab)
    const onPageHide = () => send(false);
    const onVisibilityChange = () => {
      if (document.visibilityState === "hidden") {
        send(false);
      }
    };

    video.addEventListener("loadedmetadata", tryApplyStart);
    // Some environments only become seekable around canplay.
    video.addEventListener("canplay", tryApplyStart);
    video.addEventListener("seeked", onSeeked);
    video.addEventListener("ended", onEnded);
    window.addEventListener("pagehide", onPageHide);
    document.addEventListener("visibilitychange", onVisibilityChange);

    // In case metadata is already loaded by the time we attach listeners.
    if (video.readyState >= 1) {
      tryApplyStart();
    }

    return () => {
      // Also persist on SPA navigation (component unmount).
      send(false);

      video.removeEventListener("loadedmetadata", tryApplyStart);
      video.removeEventListener("canplay", tryApplyStart);
      video.removeEventListener("seeked", onSeeked);
      video.removeEventListener("ended", onEnded);
      window.removeEventListener("pagehide", onPageHide);
      document.removeEventListener("visibilitychange", onVisibilityChange);
    };
  }, [programId, videoUrl, startPositionSeconds]);

  return (
    <div className="relative w-full">
      {/* 戻るボタン 左上 */}
      <button
        aria-label="前の画面に戻る"
        className="absolute top-3 left-3 z-20 bg-white/80 rounded-full p-1 hover:bg-gray-100 border border-gray-200 shadow"
        onClick={() => router.back()}
      >
        <svg
          width="24"
          height="24"
          viewBox="0 0 24 24"
          fill="none"
          xmlns="http://www.w3.org/2000/svg"
          className="h-6 w-6 text-gray-700"
        >
          <line x1="6" y1="6" x2="18" y2="18" stroke="currentColor" strokeWidth="2" strokeLinecap="round" />
          <line x1="18" y1="6" x2="6" y2="18" stroke="currentColor" strokeWidth="2" strokeLinecap="round" />
        </svg>
      </button>
      <video ref={ref} controls style={{ width: "100%" }} src={videoUrl}>
        お使いのブラウザは video タグに対応していません。
      </video>
    </div>
  );
};
