"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { useState } from "react";

import { authClient } from "@/lib/auth/auth-client";

export default function SignupPage() {
  const router = useRouter();

  const [name, setName] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  async function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    setIsSubmitting(true);
    setError(null);

    try {
      const res = await authClient.signUp.email({
        name,
        email,
        password,
      });

      if (res?.error) {
        setError(res.error.message ?? "登録に失敗しました");
        return;
      }

      router.push("/top");
      router.refresh();
    } catch (err) {
      setError(err instanceof Error ? err.message : "登録に失敗しました");
    } finally {
      setIsSubmitting(false);
    }
  }

  return (
    <div className="flex min-h-screen items-center justify-center font-sans dark:bg-black">
      <main className="w-full max-w-md bg-white px-10 py-12 dark:bg-black">
        <h1 className="text-3xl font-extrabold text-zinc-900 dark:text-zinc-100">会員登録</h1>
        <p className="mt-2 text-sm text-zinc-600 dark:text-zinc-300">
          すでにアカウントをお持ちですか？{" "}
          <Link className="text-blue-600" href="/login">
            ログイン
          </Link>
        </p>

        <form className="mt-8 space-y-4" onSubmit={onSubmit}>
          <label className="block">
            <span className="text-sm text-zinc-700 dark:text-zinc-200">名前</span>
            <input
              className="mt-1 w-full rounded border border-zinc-300 bg-transparent px-3 py-2 text-zinc-900 dark:border-zinc-700 dark:text-zinc-100"
              type="text"
              autoComplete="name"
              value={name}
              onChange={(e) => setName(e.target.value)}
              required
            />
          </label>

          <label className="block">
            <span className="text-sm text-zinc-700 dark:text-zinc-200">メールアドレス</span>
            <input
              className="mt-1 w-full rounded border border-zinc-300 bg-transparent px-3 py-2 text-zinc-900 dark:border-zinc-700 dark:text-zinc-100"
              type="email"
              autoComplete="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              required
            />
          </label>

          <label className="block">
            <span className="text-sm text-zinc-700 dark:text-zinc-200">パスワード</span>
            <input
              className="mt-1 w-full rounded border border-zinc-300 bg-transparent px-3 py-2 text-zinc-900 dark:border-zinc-700 dark:text-zinc-100"
              type="password"
              autoComplete="new-password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              required
            />
          </label>

          {error ? <p className="text-sm text-red-600">{error}</p> : null}

          <button
            className="w-full rounded bg-blue-600 px-4 py-2 text-white disabled:opacity-50"
            type="submit"
            disabled={isSubmitting}
          >
            {isSubmitting ? "作成中..." : "アカウント作成"}
          </button>
        </form>
      </main>
    </div>
  );
}
