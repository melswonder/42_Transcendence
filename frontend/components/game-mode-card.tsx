"use client";

import Link from "next/link";
import type { ReactNode } from "react";
import { Paper, Stack, Text, ThemeIcon } from "@mantine/core";
import { useTranslations } from "next-intl";

interface GameModeCardProps {
  icon: ReactNode;
  title: string;
  description: string;
  /** 遷移先。あるモードはリンクとして描く。 */
  href?: string;
  /** 主役のカード。1 画面に 1 枚だけ。 */
  featured?: boolean;
  /** 行き先がまだ無いモード。押せないことが見て分かる状態にする。 */
  comingSoon?: boolean;
}

export function GameModeCard({
  icon,
  title,
  description,
  href,
  featured = false,
  comingSoon = false,
}: GameModeCardProps) {
  const t = useTranslations("home");
  const body = (
    <Stack gap="sm" align="flex-start">
      <ThemeIcon size={44} radius="md" variant={featured ? "filled" : "light"}>
        {icon}
      </ThemeIcon>
      <Text fw={700} size="lg">
        {title}
      </Text>
      <Text size="sm" c="dimmed">
        {comingSoon ? t("comingSoon") : description}
      </Text>
    </Stack>
  );

  if (href && !comingSoon) {
    return (
      <Paper
        component={Link}
        href={href}
        p="lg"
        bg={featured ? "emerald.9" : undefined}
        className="block h-full text-left transition-colors hover:border-emerald-500"
      >
        {body}
      </Paper>
    );
  }

  return (
    <Paper
      component="button"
      type="button"
      p="lg"
      disabled={comingSoon}
      bg={featured ? "emerald.9" : undefined}
      className="h-full text-left transition-colors enabled:hover:border-emerald-500 disabled:opacity-50"
    >
      {body}
    </Paper>
  );
}
