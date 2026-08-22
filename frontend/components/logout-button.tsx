"use client";

import { useState } from "react";
import { useTranslations } from "next-intl";
import { ActionIcon, Button } from "@mantine/core";
import { IconLogout } from "@tabler/icons-react";

import { ConfirmModal } from "@/components/confirm-modal";
import { apiUrl } from "@/lib/api";

/** セッションを失効させてログイン画面へ戻す。押し間違い防止に確認モーダルを挟む。 */
export function LogoutButton({ iconOnly = false }: { iconOnly?: boolean }) {
  const t = useTranslations("nav");
  const [confirmOpen, setConfirmOpen] = useState(false);
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

  const trigger = iconOnly ? (
    <ActionIcon
      onClick={() => setConfirmOpen(true)}
      variant="subtle"
      color="gray"
      aria-label={t("logout")}
    >
      <IconLogout size={18} />
    </ActionIcon>
  ) : (
    <Button
      onClick={() => setConfirmOpen(true)}
      variant="subtle"
      color="gray"
      leftSection={<IconLogout size={18} />}
    >
      {t("logout")}
    </Button>
  );

  return (
    <>
      {trigger}
      <ConfirmModal
        opened={confirmOpen}
        onClose={() => setConfirmOpen(false)}
        onConfirm={logout}
        title={t("logoutConfirmTitle")}
        message={t("logoutConfirmMessage")}
        confirmLabel={t("logout")}
        color="red"
        loading={loading}
      />
    </>
  );
}
