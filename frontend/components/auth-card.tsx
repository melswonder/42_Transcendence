import type { ReactNode } from "react";
import { Paper, Stack, Text, ThemeIcon, Title } from "@mantine/core";

import { LegalLinks } from "@/components/legal-links";
import { LocaleSwitcher } from "@/components/locale-switcher";
import { QuoridorMark } from "@/components/quoridor-mark";

/** ログイン・サインアップ共通の外枠。ブランドの見出しと言語切り替えを持ち、
 * ボタンや注意書きなどの中身はページ側が children で入れる。
 */
export function AuthCard({
  tagline,
  children,
}: {
  tagline: string;
  children: ReactNode;
}) {
  return (
    <main className="flex min-h-dvh items-center justify-center bg-body p-4">
      <Paper p="xl" shadow="md" className="w-full max-w-md">
        <Stack gap="lg">
          <Stack align="center" gap="xs">
            <ThemeIcon size={64} radius="lg" variant="light">
              <QuoridorMark />
            </ThemeIcon>
            <Title order={1} className="text-3xl tracking-[0.15em]">
              TRANSCENDENCE
            </Title>
            <Text c="dimmed">{tagline}</Text>
          </Stack>

          {children}

          <LocaleSwitcher />

          <LegalLinks />
        </Stack>
      </Paper>
    </main>
  );
}
