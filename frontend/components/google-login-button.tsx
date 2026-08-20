"use client";

import { useState } from "react";
import { Button } from "@mantine/core";

import { GoogleMark } from "@/components/google-mark";
import { apiUrl } from "@/lib/api";

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
      loading={loading}
      loaderProps={{ type: "dots" }}
      fullWidth
      size="lg"
      // Google のボタンは白地が要件。theme の primaryColor ではなく
      // variant="white" を使い、文字色だけ dark に寄せる。
      variant="white"
      color="dark.7"
      leftSection={<GoogleMark />}
    >
      Google でログイン
    </Button>
  );
}
