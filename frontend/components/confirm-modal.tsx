"use client";

import type { ReactNode } from "react";
import { Button, Group, Stack, Text } from "@mantine/core";
import { useTranslations } from "next-intl";

import { AppModal } from "@/components/app-modal";

/** 確認ダイアログ。共通モーダルの中身を「メッセージ + はい/いいえ」に固定した版。
 * ログアウトや投了など、取り消しの効かない操作の前に挟む。
 */
export function ConfirmModal({
  opened,
  onClose,
  onConfirm,
  title,
  message,
  confirmLabel,
  color = "emerald",
  loading = false,
}: {
  opened: boolean;
  onClose: () => void;
  onConfirm: () => void;
  title: ReactNode;
  message?: ReactNode;
  confirmLabel: ReactNode;
  /** 破壊的な操作は "red" を渡す。 */
  color?: string;
  loading?: boolean;
}) {
  const t = useTranslations("common");

  return (
    <AppModal opened={opened} onClose={onClose} title={title}>
      <Stack gap="lg">
        {message && (
          <Text size="sm" c="dimmed">
            {message}
          </Text>
        )}
        <Group justify="flex-end" gap="sm">
          <Button variant="default" onClick={onClose} disabled={loading}>
            {t("cancel")}
          </Button>
          <Button color={color} onClick={onConfirm} loading={loading}>
            {confirmLabel}
          </Button>
        </Group>
      </Stack>
    </AppModal>
  );
}
