"use client";

import { useState } from "react";
import { Loader2Icon } from "lucide-react";

import { Button } from "@/shared/ui/button";
import { GoogleMark } from "@/shared/ui/google-mark";
import { apiUrl } from "@/shared/config/api";

export function GoogleLoginButton() {
  const [loading, setLoading] = useState(false);

  // fetch ではなく location で飛ばす。
  // OAuth は「ブラウザごと Google の同意画面へ遷移する」流れなので、
  // fetch で叩くと 302 の Location を辿るだけで同意画面を表示できない（CORS の対象にもなる）。
  const start = () => {
    setLoading(true);
    window.location.href = apiUrl("/auth/google");
  };

  return (
    <Button
      type="button"
      onClick={start}
      disabled={loading}
      size="lg"
      // Google のボタンは白地が要件なので、テーマトークンではなく固定色で上書きする。
      className="h-12 w-full gap-3 bg-white text-base font-bold text-slate-900 hover:bg-slate-100"
    >
      {loading ? (
        <>
          <Loader2Icon className="size-5 animate-spin" />
          <span className="sr-only">読み込み中</span>
        </>
      ) : (
        <>
          <GoogleMark />
          Google でログイン
        </>
      )}
    </Button>
  );
}
