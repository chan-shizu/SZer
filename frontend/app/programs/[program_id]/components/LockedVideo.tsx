"use client";

import { useRouter } from "next/navigation";

type Props = {
  thumbnailUrl: string | null;
  title: string;
};

export const LockedVideo = ({ thumbnailUrl, title }: Props) => {
  const router = useRouter();

  return (
    <div className="relative w-full aspect-video bg-black">
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
        <button className="bg-pink-600 text-white font-bold px-6 py-2.5 rounded-full shadow-lg hover:bg-pink-700 transition">
          購入する
        </button>
      </div>
    </div>
  );
};
