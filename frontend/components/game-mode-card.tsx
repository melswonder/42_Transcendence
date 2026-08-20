import type { ReactNode } from "react";
import { Paper, Stack, Text, ThemeIcon } from "@mantine/core";

interface GameModeCardProps {
  icon: ReactNode;
  title: string;
  description: string;
  /** 主役のカード。1 画面に 1 枚だけ。 */
  featured?: boolean;
  /** 行き先がまだ無いモード。押せないことが見て分かる状態にする。 */
  comingSoon?: boolean;
}

export function GameModeCard({
  icon,
  title,
  description,
  featured = false,
  comingSoon = false,
}: GameModeCardProps) {
  return (
    <Paper
      component="button"
      type="button"
      p="lg"
      disabled={comingSoon}
      bg={featured ? "emerald.9" : undefined}
      className="h-full text-left transition-colors enabled:hover:border-emerald-500 disabled:opacity-50"
    >
      <Stack gap="sm" align="flex-start">
        <ThemeIcon
          size={44}
          radius="md"
          variant={featured ? "filled" : "light"}
        >
          {icon}
        </ThemeIcon>
        <Text fw={700} size="lg">
          {title}
        </Text>
        <Text size="sm" c="dimmed">
          {comingSoon ? "準備中" : description}
        </Text>
      </Stack>
    </Paper>
  );
}
