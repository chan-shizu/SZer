"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { SimpleModal } from "@/components/SimpleModal";
import { AuthModal } from "@/components/AuthModal";
import { authClient } from "@/lib/auth/auth-client";
import { createPayPayCheckout } from "@/lib/api/paypay";

type Props = {
  thumbnailUrl: string | null;
  title: string;
  price: number;
  programId: number;
};

export const LockedVideo = ({ thumbnailUrl, title, price, programId }: Props) => {
  const router = useRouter();
  const [showModal, setShowModal] = useState(false);
  const [showAuthModal, setShowAuthModal] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState("");

  const openPurchaseModal = async () => {
    const { data: session } = await authClient.getSession();
    if (!session?.user) {
      setShowAuthModal(true);
      return;
    }
    setError("");
    setShowModal(true);
  };

  const handlePurchase = async () => {
    setIsSubmitting(true);
    setError("");
    try {
      const res = await createPayPayCheckout(programId);
      const url = res.url || res.deeplink;
      if (!url) {
        throw new Error("PayPay checkout URL is missing");
      }
      window.location.href = url;
    } catch (e) {
      const msg = e instanceof Error ? e.message : "";
      if (msg.includes("already purchased")) {
        setShowModal(false);
        router.refresh();
      } else {
        setError("決済の開始に失敗しました。もう一度お試しください。");
      }
      setIsSubmitting(false);
    }
  };

  return (
    <>
      <div className="relative w-full aspect-video bg-black">
        {/* 戻るボタン 左上 */}
        <button
          aria-label="前の画面に戻る"
          className="absolute top-3 left-3 z-20 bg-white/80 rounded-full p-1 hover:bg-subtle border border-border shadow"
          onClick={() => router.back()}
        >
          <svg
            width="24"
            height="24"
            viewBox="0 0 24 24"
            fill="none"
            xmlns="http://www.w3.org/2000/svg"
            className="h-6 w-6 text-muted-foreground"
          >
            <line x1="6" y1="6" x2="18" y2="18" stroke="currentColor" strokeWidth="2" strokeLinecap="round" />
            <line x1="18" y1="6" x2="6" y2="18" stroke="currentColor" strokeWidth="2" strokeLinecap="round" />
          </svg>
        </button>
        {thumbnailUrl && (
          /* eslint-disable-next-line @next/next/no-img-element */
          <img
            src={thumbnailUrl}
            alt={title}
            className="absolute inset-0 w-full h-full object-cover blur-xl scale-110"
          />
        )}
        <div className="absolute inset-0 bg-black/40" />
        <div className="absolute inset-0 flex flex-col items-center justify-center z-10">
          <svg
            xmlns="http://www.w3.org/2000/svg"
            viewBox="0 0 24 24"
            fill="currentColor"
            className="w-10 h-10 text-white/80 mb-3"
          >
            <path
              fillRule="evenodd"
              d="M12 1.5a5.25 5.25 0 0 0-5.25 5.25v3a3 3 0 0 0-3 3v6.75a3 3 0 0 0 3 3h10.5a3 3 0 0 0 3-3v-6.75a3 3 0 0 0-3-3v-3A5.25 5.25 0 0 0 12 1.5Zm3.75 8.25v-3a3.75 3.75 0 1 0-7.5 0v3h7.5Z"
              clipRule="evenodd"
            />
          </svg>
          <span className="text-white font-bold text-lg mb-1">この動画は購入者限定です</span>
          <span className="text-white/80 text-sm mb-4">購入すると視聴できるようになります</span>
          <button
            className="bg-brand text-white font-bold px-6 py-2.5 rounded-full shadow-lg hover:bg-orange-700 transition"
            onClick={openPurchaseModal}
          >
            購入する
          </button>
        </div>
      </div>

      <SimpleModal open={showModal} onClose={() => setShowModal(false)}>
        <div className="flex flex-col items-center gap-y-4">
          <h2 className="text-lg font-bold">番組を購入</h2>
          <p className="text-sm text-muted-foreground text-center">{title}</p>
          <p className="text-2xl font-bold text-brand">
            ¥{price.toLocaleString()}（税込）
          </p>
          {error && (
            <p className="text-sm text-red-600 text-center">{error}</p>
          )}
          <button
            className="w-full bg-brand text-white font-bold py-3 rounded-lg hover:bg-orange-700 transition disabled:opacity-50"
            onClick={handlePurchase}
            disabled={isSubmitting}
          >
            {isSubmitting ? "処理中..." : "PayPayで支払う"}
          </button>
          <button
            className="w-full text-sm text-muted-foreground py-2"
            onClick={() => setShowModal(false)}
            disabled={isSubmitting}
          >
            キャンセル
          </button>
        </div>
      </SimpleModal>

      <AuthModal open={showAuthModal} onClose={() => setShowAuthModal(false)} />
    </>
  );
};
