"use client";

import { useState } from "react";

import { addPoints } from "@/lib/api/points";
import { createPayPayCheckout } from "@/lib/api/paypay";

type Amount = 100 | 500 | 1000;
const AMOUNTS: Amount[] = [100, 500, 1000];

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

  async function onPayPayBuy(amountYen: Amount) {
    setIsSubmitting(true);
    setError(null);

    try {
      const res = await createPayPayCheckout(amountYen);
      const url = res.url || res.deeplink;
      if (!url) {
        throw new Error("paypay checkout url is missing");
      }
      window.location.href = url;
    } catch (e) {
      setError(e instanceof Error ? e.message : "failed to start paypay checkout");
    } finally {
      setIsSubmitting(false);
    }
  }

  return (
    <div className="space-y-4 px-4">
      <div className="border-b border-muted pb-3">
        <p className="text-xs text-muted-foreground">現在のポイント</p>
        <p className="mt-1 text-sm text-foreground">{currentPoints.toLocaleString()} pt</p>
      </div>

      <div className="space-y-2">
        <p className="text-xs text-muted-foreground">PayPayで購入（1円=1pt）</p>
        {AMOUNTS.map((amount) => (
          <button
            key={amount}
            type="button"
            disabled={isSubmitting}
            onClick={() => onPayPayBuy(amount)}
            className="w-full rounded border border-muted px-4 py-3 text-sm text-foreground disabled:opacity-50"
          >
            {amount.toLocaleString()} 円で購入
          </button>
        ))}
      </div>

      {error && <p className="text-sm text-red-500">{error}</p>}
    </div>
  );
}
