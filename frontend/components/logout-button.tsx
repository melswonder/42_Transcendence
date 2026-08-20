"use client";

import { useState } from "react";
import { ActionIcon, Button } from "@mantine/core";
import { IconLogout } from "@tabler/icons-react";

import { apiUrl } from "@/lib/api";

/** セッションを失効させてログイン画面へ戻す。 */
export function LogoutButton({ iconOnly = false }: { iconOnly?: boolean }) {
  const [loading, setLoading] = useState(false);

  const logout = async () => {
    setLoading(true);
    try {
      // credentials を付けないと Cookie が飛ばず、サーバー側のセッションが残る。
      await fetch(apiUrl("/auth/logout"), {
        method: "POST",
        credentials: "include",
      });
      // router.refresh() ではなく遷移させる。Server Component が握っている
      // ユーザー情報ごと捨てたいため。
      window.location.href = "/login";
    } catch {
      setLoading(false);
    }
  };

  if (iconOnly) {
    return (
      <ActionIcon
        onClick={logout}
        loading={loading}
        variant="subtle"
        color="gray"
        aria-label="ログアウト"
      >
        <IconLogout size={18} />
      </ActionIcon>
    );
  }

  return (
    <Button
      onClick={logout}
      loading={loading}
      variant="subtle"
      color="gray"
      leftSection={<IconLogout size={18} />}
    >
      ログアウト
    </Button>
  );
}
