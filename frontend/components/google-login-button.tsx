"use client";

import { useState } from "react";
import { useTranslations } from "next-intl";
import { Button } from "@mantine/core";

import { GoogleMark } from "@/components/google-mark";
import { apiUrl } from "@/lib/api";

/** Google OAuth の開始ボタン。サインアップ画面では文言だけ変える
 * （フローは同じで、初回ログイン時にアカウントが作られる）。
 */
export function GoogleLoginButton({
  mode = "login",
}: {
  mode?: "login" | "signup";
}) {
  const t = useTranslations(mode);
  const [loading, setLoading] = useState(false);

  // fetch にしないこと。OAuth はブラウザごと同意画面へ遷移する流れなので、
  // fetch だと 302 の Location を辿るだけで同意画面を出せない。
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
      // 白地は Google のブランド要件。theme の primaryColor は使わない。
      variant="white"
      color="dark.7"
      leftSection={<GoogleMark />}
    >
      {t("googleButton")}
    </Button>
  );
}
