"use client";

import Link from "next/link";
import type { ReactNode } from "react";
import { Paper, Stack, Text, ThemeIcon } from "@mantine/core";

interface GameModeCardProps {
  icon: ReactNode;
  title: string;
  description: string;
  /** 遷移先。カード全体がリンクになる。 */
  href: string;
  /** 主役のカード。1 画面に 1 枚だけ。 */
  featured?: boolean;
}

export function GameModeCard({
  icon,
  title,
  description,
  href,
  featured = false,
}: GameModeCardProps) {
  const body = (
    <Stack gap="sm" align="flex-start">
      <ThemeIcon size={44} radius="md" variant={featured ? "filled" : "light"}>
        {icon}
      </ThemeIcon>
      <Text fw={700} size="lg">
        {title}
      </Text>
      <Text size="sm" c="dimmed">
        {description}
      </Text>
    </Stack>
  );

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
