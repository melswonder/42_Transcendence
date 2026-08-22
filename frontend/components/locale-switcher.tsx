"use client";

import { useRouter } from "next/navigation";
import { useState } from "react";
import { SegmentedControl } from "@mantine/core";
import { useLocale } from "next-intl";

import { localeCookie } from "@/i18n/config";
import { apiUrl } from "@/lib/api";

/** 言語スイッチャー。
 * Cookie に保存するのでリロードしても保たれる。ログイン中なら
 * users.preferred_locale にも書いて、別ブラウザでも設定画面から追える。
 */
export function LocaleSwitcher({ loggedIn = false }: { loggedIn?: boolean }) {
  const router = useRouter();
  const locale = useLocale();
  const [value, setValue] = useState(locale);

  const change = (next: string) => {
    setValue(next);
    document.cookie = `${localeCookie}=${next}; path=/; max-age=31536000; samesite=lax`;
    if (loggedIn) {
      // 失敗しても Cookie だけで言語は切り替わるので握りつぶす。
      void fetch(apiUrl("/users/me"), {
        method: "PATCH",
        credentials: "include",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ preferred_locale: next }),
      }).catch(() => {});
    }
    router.refresh();
  };

  return (
    <SegmentedControl
      size="xs"
      fullWidth
      value={value}
      onChange={change}
      data={[
        { value: "ja", label: "日本語" },
        { value: "en", label: "EN" },
        { value: "fr", label: "FR" },
      ]}
    />
  );
}
