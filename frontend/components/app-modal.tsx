"use client";

import type { ReactNode } from "react";
import { Modal, Text } from "@mantine/core";

/** アプリ共通のモーダル。
 * 見た目（中央寄せ・角丸・オーバーレイ）をここで揃え、中身は children で差し替える。
 * 勝敗の表示や確認ダイアログなど、種類が増えてもこの殻だけを使い回す。
 */
export function AppModal({
  opened,
  onClose,
  title,
  children,
  size = "sm",
  closeable = true,
}: {
  opened: boolean;
  onClose: () => void;
  title?: ReactNode;
  children: ReactNode;
  size?: string | number;
  /** 背景クリックや Esc で閉じてよいか。結果表示などは true のままでよい。 */
  closeable?: boolean;
}) {
  return (
    <Modal
      opened={opened}
      onClose={onClose}
      title={
        title && (
          <Text fw={700} size="lg" component="span">
            {title}
          </Text>
        )
      }
      size={size}
      centered
      radius="md"
      withCloseButton={closeable}
      closeOnClickOutside={closeable}
      closeOnEscape={closeable}
      overlayProps={{ backgroundOpacity: 0.6, blur: 2 }}
    >
      {children}
    </Modal>
  );
}
