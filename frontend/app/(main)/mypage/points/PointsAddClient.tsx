'use client';

import { useState } from "react";

import { addPoints } from "@/lib/api/points";

type Amount = 100 | 500 | 1000;

export function PointsAddClient({ initialPoints }: { initialPoints: number }) {
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [currentPoints, setCurrentPoints] = useState<number>(initialPoints);
  const [error, setError] = useState<string | null>(null);

  async function onAdd(amount: Amount) {
    setIsSubmitting(true);
    setError(null);

    try {
      const res = await addPoints(amount);
      setCurrentPoints(res.points);
    } catch (e) {
      setError(e instanceof Error ? e.message : "failed to add points");
    } finally {
      setIsSubmitting(false);
    }
  }

  return (
    <div className="space-y-3 p-3">
      <div className="space-y-2">
        <button
          type="button"
          disabled={isSubmitting}
          onClick={() => onAdd(100)}
          className="w-full rounded-md border border-muted px-4 py-3 text-sm font-medium text-foreground disabled:opacity-50"
        >
          100ポイント追加
        </button>
        <button
          type="button"
          disabled={isSubmitting}
          onClick={() => onAdd(500)}
          className="w-full rounded-md border border-muted px-4 py-3 text-sm font-medium text-foreground disabled:opacity-50"
        >
          500ポイント追加
        </button>
        <button
          type="button"
          disabled={isSubmitting}
          onClick={() => onAdd(1000)}
          className="w-full rounded-md border border-muted px-4 py-3 text-sm font-medium text-foreground disabled:opacity-50"
        >
          1000ポイント追加
        </button>
      </div>

      <div className="text-sm text-foreground">現在のポイント: {currentPoints}</div>

      {error ? <div className="text-sm text-muted-foreground">{error}</div> : null}
    </div>
  );
}
