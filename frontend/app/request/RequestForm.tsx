"use client";

import { useState } from "react";
import { postRequestClient } from "@/lib/api/requests.client";

export default function RequestForm() {
  const [content, setContent] = useState("");
  const [name, setName] = useState("");
  const [contact, setContact] = useState("");
  const [note, setNote] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [submitted, setSubmitted] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!content.trim() || !name.trim() || !contact.trim()) return;

    setLoading(true);
    setError(null);
    try {
      await postRequestClient({ content, name, contact, note });
      setContent("");
      setName("");
      setContact("");
      setNote("");
      setSubmitted(true);
    } catch {
      setError("リクエストの送信に失敗しました。もう一度お試しください。");
    } finally {
      setLoading(false);
    }
  };

  if (submitted) {
    return (
      <div className="text-center py-12">
        <p className="text-lg font-semibold mb-4">リクエストを送信しました！</p>
        <p className="text-sm text-gray-500 mb-6">ご要望ありがとうございます。</p>
        <button onClick={() => setSubmitted(false)} className="px-6 py-2 bg-black text-white rounded-lg text-sm">
          続けてリクエストする
        </button>
      </div>
    );
  }

  return (
    <form onSubmit={handleSubmit} className="flex flex-col gap-4">
      <div>
        <label htmlFor="content" className="block text-sm font-medium mb-1">
          リクエスト内容 <span className="text-red-500">*</span>
        </label>
        <textarea
          id="content"
          className="w-full bg-gray-100 rounded-lg px-4 py-2 text-base min-h-[100px] resize-y"
          placeholder="見たい番組や出演者など..."
          value={content}
          onChange={(e) => setContent(e.target.value)}
          disabled={loading}
          maxLength={1000}
        />
      </div>

      <div>
        <label htmlFor="name" className="block text-sm font-medium mb-1">
          お名前
          <span className="text-xs font-normal">(公開されません、僕が連絡取れるように)</span>
          <span className="text-red-500">*</span>
        </label>
        <input
          id="name"
          type="text"
          className="w-full bg-gray-100 rounded-lg px-4 py-2 text-base"
          placeholder="お名前"
          value={name}
          onChange={(e) => setName(e.target.value)}
          disabled={loading}
          maxLength={100}
        />
      </div>

      <div>
        <label htmlFor="contact" className="block text-sm font-medium mb-1">
          連絡先
          <span className="text-xs font-normal">(既に連絡先交換している方はその旨書いてもらえれば！)</span>
          <span className="text-red-500">*</span>
        </label>
        <input
          id="contact"
          type="text"
          className="w-full bg-gray-100 rounded-lg px-4 py-2 text-base"
          placeholder="インスタのDMで、example@example.com"
          value={contact}
          onChange={(e) => setContact(e.target.value)}
          disabled={loading}
          maxLength={200}
        />
      </div>

      <div>
        <label htmlFor="note" className="block text-sm font-medium mb-1">
          備考
        </label>
        <textarea
          id="note"
          className="w-full bg-gray-100 rounded-lg px-4 py-2 text-base min-h-[60px] resize-y"
          placeholder="その他ご要望があれば..."
          value={note}
          onChange={(e) => setNote(e.target.value)}
          disabled={loading}
          maxLength={1000}
        />
      </div>

      {error && <p className="text-red-500 text-sm">{error}</p>}

      <button
        type="submit"
        className="w-full py-3 bg-black text-white rounded-lg text-sm font-medium disabled:opacity-50 transition"
        disabled={loading || !content.trim() || !name.trim() || !contact.trim()}
      >
        {loading ? "送信中..." : "リクエストを送信"}
      </button>
    </form>
  );
}
