"use client";

import { useState } from "react";

// バックエンドの入口。docker-compose が NEXT_PUBLIC_API_URL を渡している。
// フロント自身（:3000）ではなくバックエンド（:4000）を指すこと。
// 末尾スラッシュ付きで設定されても "//auth/google" にならないよう落としておく。
const API_BASE = (
  process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:4000"
).replace(/\/+$/, "");

export function GoogleLoginButton() {
  const [loading, setLoading] = useState(false);

  // fetch ではなく location で飛ばす。
  // OAuth は「ブラウザごと Google の同意画面へ遷移する」流れなので、
  // fetch で叩くと 302 の Location を辿るだけで同意画面を表示できない（CORS の対象にもなる）。
  const start = () => {
    setLoading(true);
    window.location.href = `${API_BASE}/auth/google`;
  };

  return (
    <button
      type="button"
      onClick={start}
      disabled={loading}
      className="flex h-12 w-full items-center justify-center gap-3 rounded-lg bg-white font-bold text-slate-900 transition-colors hover:bg-slate-100 disabled:cursor-not-allowed disabled:opacity-70"
    >
      {loading ? (
        <span
          className="h-6 w-6 animate-spin rounded-full border-2 border-slate-900 border-t-transparent"
          aria-label="読み込み中"
        />
      ) : (
        <>
          <GoogleMark />
          Google でログイン
        </>
      )}
    </button>
  );
}

// Google ブランドの G マーク（公式の 4 色）。
function GoogleMark() {
  return (
    <svg width="20" height="20" viewBox="0 0 48 48" aria-hidden="true">
      <path
        fill="#4285F4"
        d="M45.12 24.5c0-1.56-.14-3.06-.4-4.5H24v8.51h11.84c-.51 2.75-2.06 5.08-4.39 6.64v5.52h7.11c4.16-3.83 6.56-9.47 6.56-16.17z"
      />
      <path
        fill="#34A853"
        d="M24 46c5.94 0 10.92-1.97 14.56-5.33l-7.11-5.52c-1.97 1.32-4.49 2.1-7.45 2.1-5.73 0-10.58-3.87-12.31-9.07H4.34v5.7C7.96 41.07 15.4 46 24 46z"
      />
      <path
        fill="#FBBC05"
        d="M11.69 28.18C11.25 26.86 11 25.45 11 24s.25-2.86.69-4.18v-5.7H4.34C2.85 17.09 2 20.45 2 24s.85 6.91 2.34 9.88l7.35-5.7z"
      />
      <path
        fill="#EA4335"
        d="M24 10.75c3.23 0 6.13 1.11 8.41 3.29l6.31-6.31C34.91 4.18 29.93 2 24 2 15.4 2 7.96 6.93 4.34 14.12l7.35 5.7c1.73-5.2 6.58-9.07 12.31-9.07z"
      />
    </svg>
  );
}
