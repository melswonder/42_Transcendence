import type { ReactNode } from "react";
import Image from "next/image";
import { Paper, Stack, Text, Title } from "@mantine/core";
import { useTranslations } from "next-intl";

import { LegalLinks } from "@/components/legal-links";
import { LocaleSwitcher } from "@/components/locale-switcher";

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
  const t = useTranslations("common");

  return (
    <main className="flex min-h-dvh items-center justify-center bg-body p-4">
      <Paper p="xl" shadow="md" className="w-full max-w-md">
        <Stack gap="lg">
          <Stack align="center" gap="xs">
            <Image
              src="/icon.png"
              alt=""
              width={64}
              height={64}
              className="rounded-2xl"
            />
            <Title order={1} className="text-3xl tracking-[0.15em]">
              {t("appName")}
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
