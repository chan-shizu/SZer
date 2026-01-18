"use client";

import { useEffect, useRef } from "react";

import { upsertWatchHistory, upsertWatchHistoryBeacon } from "@/lib/api/watch_histories";

type Props = {
  programId: number;
  videoUrl: string;
  startPositionSeconds?: number;
};

export const Video = ({ programId, videoUrl, startPositionSeconds }: Props) => {
  const ref = useRef<HTMLVideoElement | null>(null);
  const sentRef = useRef(false);
  const appliedStartRef = useRef(false);

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

    const send = (isCompleted: boolean) => {
      if (sentRef.current) return;
      const positionSeconds = Math.floor(video.currentTime || 0);
      if (positionSeconds < 1.0 && !isCompleted) {
        // do not send very beginning position to reduce API calls
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
    <video ref={ref} controls style={{ width: "100%", maxWidth: 960 }} src={videoUrl}>
      お使いのブラウザは video タグに対応していません。
    </video>
  );
};
